// Copyright (c) 2024 ExonLabs, All rights reserved.
// Use of this source code is governed by a BSD 3-Clause
// license that can be found in the LICENSE file.

package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
	"time"

	"github.com/exonlabs/go-utils/pkg/abc/dictx"
	"github.com/exonlabs/go-utils/pkg/comm"
	"github.com/exonlabs/go-utils/pkg/comm/netcomm"
	"github.com/exonlabs/go-utils/pkg/comm/serialcomm"
	"github.com/exonlabs/go-utils/pkg/comm/sockcomm"
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
MIIB/DCCAaGgAwIBAgIUYckcHG8RtlTQXJbkpMpbKEeHe6kwCgYIKoZIzj0EAwIw
MTELMAkGA1UEBhMCRUcxETAPBgNVBAoMCEV4b25MYWJzMQ8wDQYDVQQDDAZSb290
Q0EwHhcNMjQxMTIwMTM0MjQ5WhcNMzQxMTE4MTM0MjQ5WjAyMQswCQYDVQQGEwJF
RzERMA8GA1UECgwIRXhvbkxhYnMxEDAOBgNVBAMMB2NsaWVudDEwWTATBgcqhkjO
PQIBBggqhkjOPQMBBwNCAASZUGX5RieM62qhM5XKwo9uICoX/kxvhYmhV4jYof6O
KmaSIuw7r/TeELPs362o5SbTrbh6m+Fcn1JPFugTBbLxo4GVMIGSMA4GA1UdDwEB
/wQEAwIDqDATBgNVHSUEDDAKBggrBgEFBQcDAjArBgNVHREEJDAigg1jbGllbnQx
LmxvY2FsggtjbGllbnQxLmxhbocEfwAAATAdBgNVHQ4EFgQUPzd45v+ZOvm2Rydu
EDJbRFJbIKUwHwYDVR0jBBgwFoAU1DaM6ZHuy2fYEbV0ivm65f6/HlgwCgYIKoZI
zj0EAwIDSQAwRgIhANq2eaZA2BCy166GQt+yyDeG1AmBhspA/doxLf8cusJgAiEA
tyV1nBIkDlJ5JHCZ+DPqnc/v5Kgk9xICJO6CHgkafcM=
-----END CERTIFICATE-----`
	LocalKey = `-----BEGIN EC PARAMETERS-----
BggqhkjOPQMBBw==
-----END EC PARAMETERS-----
-----BEGIN EC PRIVATE KEY-----
MHcCAQEEIPhTCAa+nUofAYH/TRjym3/fMAx2x8nYSzUsLKPYFb3WoAoGCCqGSM49
AwEHoUQDQgAEmVBl+UYnjOtqoTOVysKPbiAqF/5Mb4WJoVeI2KH+jipmkiLsO6/0
3hCz7N+tqOUm0624epvhXJ9STxboEwWy8Q==
-----END EC PRIVATE KEY-----`
)

func run(cli comm.Connection) {
	if err := cli.Open(-1); err != nil {
		if errors.Is(err, comm.ErrClosed) || errors.Is(err, comm.ErrBreak) {
			return
		}
		panic(err)
	}
	defer cli.Close()

	// // Test receving hello msg at start of connection
	// if data, err := cli.Recv(3); err == nil {
	// 	fmt.Printf("received: %d bytes  %s", len(data), string(data))
	// 	fmt.Println("--------------------------------")
	// 	time.Sleep(500 * time.Millisecond)
	// }

	num_of_msgs := 9
	for i := 1; i <= num_of_msgs; i++ {
		msg := []byte(fmt.Sprintf("msg: %d\n", i))
		fmt.Printf("sending: %d bytes  %s", len(msg), string(msg))
		err := cli.Send(msg, -1)
		if err == nil {
			data, err := cli.Recv(3)
			if err == nil {
				fmt.Printf("received: %d bytes  %s", len(data), string(data))
				fmt.Println("--------------------------------")
			}
		}
		if err != nil {
			if errors.Is(err, comm.ErrClosed) {
				fmt.Println("connection closed")
				return
			} else {
				fmt.Println("error:", err)
			}
		}

		// add delay between messages
		if i < num_of_msgs {
			time.Sleep(500 * time.Millisecond)
		}
	}

	fmt.Println("end connection")
}

func main() {
	fmt.Printf("\n**** starting ****\n")

	com := "/dev/ttyUSB1"
	if runtime.GOOS == "windows" {
		com = "COM2"
	}
	sock := filepath.Join(os.TempDir(), "comm_sock")
	uri := flag.String(
		"uri", "", "connection uri\n"+
			"tcp:     tcp@127.0.0.1:1234\n"+
			"sock:    sock@"+sock+"\n"+
			"serial:  serial@"+com+":115200:8N1\n")
	tls := flag.Bool(
		"tls", false, "use encrypted TLS for TCP connections")
	mtls := flag.Bool(
		"mtls", false, "use Mutual-TLS authentication for TCP connections")
	flag.Parse()

	commLog := logging.NewStdoutLogger("comm")
	commLog.SetFormatter(logging.RawFormatter)

	// optional args
	opts := dictx.Dict{
		"poll_timeout":   0.01,
		"poll_chunksize": 4096,
		"poll_maxsize":   1048576,
		// "keepalive_interval": 30,
	}

	// TLS config
	if *tls {
		dictx.Merge(opts, dictx.Dict{
			"tls_enable":      true,
			"tls_min_version": 1.2,
			"tls_max_version": 1.3,
			"tls_ca_certs":    RootCrt,
			"tls_local_cert":  LocalCrt,
			"tls_local_key":   LocalKey,
		})
	}
	if *mtls {
		dictx.Merge(opts, dictx.Dict{
			"tls_mutual_auth": true,
			"tls_server_name": "server1.local",
		})
	}

	var cli comm.Connection
	var err error

	// Determine the connection type from the URI prefix
	switch strings.ToLower(strings.SplitN(*uri, "@", 2)[0]) {
	case "tcp", "tcp4", "tcp6", "udp", "udp4", "udp6":
		cli, err = netcomm.NewConnection(*uri, commLog, opts)
	case "sock":
		cli, err = sockcomm.NewConnection(*uri, commLog, opts)
	case "serial":
		cli, err = serialcomm.NewConnection(*uri, commLog, opts)
	default:
		fmt.Printf("\nError: invalid uri type\n\n")
		return
	}
	if err != nil {
		fmt.Printf("\nError: %s\n\n", err.Error())
		return
	}

	// register callback for close signals
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh,
		syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGHUP)
	go func() {
		s := <-sigCh
		fmt.Printf("\nreceived signal: %s\n", s)
		cli.Close()
	}()

	run(cli)

	fmt.Printf("\n**** exit ****\n\n")
}
