// Copyright (c) 2024 ExonLabs, All rights reserved.
// Use of this source code is governed by a BSD 3-Clause
// license that can be found in the LICENSE file.

package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"

	"github.com/exonlabs/go-utils/pkg/abc/dictx"
	"github.com/exonlabs/go-utils/pkg/comm"
	"github.com/exonlabs/go-utils/pkg/comm/commutils"
	"github.com/exonlabs/go-utils/pkg/logging"
)

var (
	RootCrt = `-----BEGIN CERTIFICATE-----
MIIBxzCCAW2gAwIBAgIUUUCRgr2OJ8tLk++YLFVXjvcPybEwCgYIKoZIzj0EAwIw
MTELMAkGA1UEBhMCRUcxETAPBgNVBAoMCEV4b25MYWJzMQ8wDQYDVQQDDAZSb290
Q0EwHhcNMjQxMTIwMTM0MjQ5WhcNMzQxMTE4MTM0MjQ5WjAxMQswCQYDVQQGEwJF
RzERMA8GA1UECgwIRXhvbkxhYnMxDzANBgNVBAMMBlJvb3RDQTBZMBMGByqGSM49
AgEGCCqGSM49AwEHA0IABKx6c9V44c4b39I1ry7lEYxNAuq+k2iyzgyANp1uylZV
L2kLydMjDsJNq7xlX2iJw01EXXRNVmp8TWl8gR+wpTKjYzBhMB0GA1UdDgQWBBTU
Nozpke7LZ9gRtXSK+brl/r8eWDAfBgNVHSMEGDAWgBTUNozpke7LZ9gRtXSK+brl
/r8eWDAPBgNVHRMBAf8EBTADAQH/MA4GA1UdDwEB/wQEAwIBBjAKBggqhkjOPQQD
AgNIADBFAiEAyIM3ntXiYgvp7mBTmgGm2lY2FYv2zEM74DDGcPZRvK0CIA8zyCRq
PdJ+CsdFfChkRkpp+YyWgtmM1azxOKy9no6l
-----END CERTIFICATE-----`
	LocalCrt = `-----BEGIN CERTIFICATE-----
MIIB+zCCAaGgAwIBAgIUO2/jyAIUEZqtc6Zk3wT86+pi5UQwCgYIKoZIzj0EAwIw
MTELMAkGA1UEBhMCRUcxETAPBgNVBAoMCEV4b25MYWJzMQ8wDQYDVQQDDAZSb290
Q0EwHhcNMjQxMTIwMTM0MjQ5WhcNMzQxMTE4MTM0MjQ5WjAyMQswCQYDVQQGEwJF
RzERMA8GA1UECgwIRXhvbkxhYnMxEDAOBgNVBAMMB3NlcnZlcjEwWTATBgcqhkjO
PQIBBggqhkjOPQMBBwNCAASHKsMfTPh5mebyGIyXhlJkQ4ROX7/nlp4rwwEq2TUA
k5rtVVy7TEuJBflBcdqibVBMWtLedLD3dKGwXStBrmtAo4GVMIGSMA4GA1UdDwEB
/wQEAwIDqDATBgNVHSUEDDAKBggrBgEFBQcDATArBgNVHREEJDAigg1zZXJ2ZXIx
LmxvY2FsggtzZXJ2ZXIxLmxhbocEfwAAATAdBgNVHQ4EFgQUUKhBtuLFDjlXfLxA
VFlWtNYwIOQwHwYDVR0jBBgwFoAU1DaM6ZHuy2fYEbV0ivm65f6/HlgwCgYIKoZI
zj0EAwIDSAAwRQIgSjHw7We2GsQVEfJlCoebYWO3uAkr+tz10pow+RV1U40CIQCI
+9GK1r7tIpyGNYy6lqP7+K0acpihlb6R5un2pJf6Ow==
-----END CERTIFICATE-----`
	LocalKey = `-----BEGIN EC PARAMETERS-----
BggqhkjOPQMBBw==
-----END EC PARAMETERS-----
-----BEGIN EC PRIVATE KEY-----
MHcCAQEEIHlD/wUdE8oxFBvwX+Lt1s5UykIrFt2DtoH7Fp4wppS1oAoGCCqGSM49
AwEHoUQDQgAEhyrDH0z4eZnm8hiMl4ZSZEOETl+/55aeK8MBKtk1AJOa7VVcu0xL
iQX5QXHaom1QTFrS3nSw93ShsF0rQa5rQA==
-----END EC PRIVATE KEY-----`
)

func HandleConnection(conn comm.Connection) {
	switch conn.Type() {
	case "tcp", "tcp4", "tcp6":
		defer fmt.Println("End peer connection", conn)
		fmt.Println("New peer connection", conn)

		// // send hello msg at start of connection
		// if err := conn.Send([]byte("HELLO\n"), 0); err != nil {
		// 	fmt.Println(err.Error())
		// 	return
		// }
	}

	for conn.IsOpened() {
		data, addr, err := conn.RecvFrom(0)
		if err != nil {
			if conn.IsOpened() && err != comm.ErrClosed {
				fmt.Println("error receiving:", err)
				continue
			} else {
				break
			}
		}

		msg := strings.TrimSpace(string(data))
		if addr == nil {
			fmt.Println("received:", msg)
		} else {
			fmt.Printf("received: %s  (%v)\n", msg, addr)
		}

		switch msg {
		case "STOP_PEER":
			conn.SendTo([]byte("peer stopped by server"), addr, 0)
			conn.Close()
			return
		case "STOP_SERVER":
			conn.SendTo([]byte("server stopped"), addr, 0)
			conn.Parent().Stop()
			return
		default:
			err := conn.SendTo([]byte("echo: "+msg+"\n"), addr, 0)
			if err != nil {
				fmt.Println("error:", err)
			}
		}
		fmt.Println("--------------------------------")
	}
}

func main() {
	fmt.Printf("\n**** starting ****\n")

	com := "/dev/ttyUSB0"
	if runtime.GOOS == "windows" {
		com = "COM1"
	}
	sock := filepath.Join(os.TempDir(), "comm_sock")
	uri := flag.String(
		"uri", "", "connection uri\n"+
			"tcp:     tcp@0.0.0.0:1234\n"+
			"sock:    sock@"+sock+"\n"+
			"serial:  serial@"+com+":115200:8N1\n")
	multi := flag.Bool(
		"multi", false, "allow multiple sessions for TCP connections")
	tls := flag.Bool(
		"tls", false, "use encrypted TLS for TCP connections")
	mtls := flag.Bool(
		"mtls", false, "use Mutual-TLS authentication for TCP connections")
	flag.Parse()

	commLog := logging.NewStdoutLogger("comm")

	// optional args
	opts := dictx.Dict{
		"poll_timeout":      0.01,
		"poll_chunksize":    4096,
		"poll_maxsize":      1048576,
		"connections_limit": 1,
	}
	if *multi {
		dictx.Set(opts, "connections_limit", 5)
	}

	// TLS config
	if *tls {
		// just for testing file based crt and key files
		crtpath := filepath.Join(os.TempDir(), "crt.pem")
		keypath := filepath.Join(os.TempDir(), "key.pem")
		os.WriteFile(crtpath, []byte(LocalCrt), 0o664)
		os.WriteFile(keypath, []byte(LocalKey), 0o664)
		defer func() {
			os.Remove(crtpath)
			os.Remove(keypath)
		}()

		dictx.Merge(opts, dictx.Dict{
			"tls_enable":      true,
			"tls_min_version": 1.2,
			"tls_max_version": 1.3,
			"tls_ca_certs":    RootCrt,
			"tls_local_cert":  crtpath,
			"tls_local_key":   keypath,
		})
	}
	if *mtls {
		dictx.Set(opts, "tls_mutual_auth", true)
	}

	srv, err := commutils.NewListener(*uri, commLog, opts)
	if err != nil {
		panic(err)
	}
	srv.ConnectionHandler(HandleConnection)

	// register callback for close signals
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh,
		syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGHUP)
	go func() {
		for {
			s := <-sigCh
			fmt.Printf("\nreceived signal: %s\n", s)
			srv.Stop()
		}
	}()

	if err := srv.Start(); err != nil {
		panic(err)
	}

	fmt.Printf("\n**** exit ****\n\n")
}
