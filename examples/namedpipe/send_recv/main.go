// Copyright (c) 2024 ExonLabs, All rights reserved.
// Use of this source code is governed by a BSD 3-Clause
// license that can be found in the LICENSE file.

//go:build !windows

package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/exonlabs/go-utils/pkg/unix/namedpipes"
)

// Description:
//
// This example demonstrates inter-process communication using named pipes.
// The Sender and Receiver structs simulate a message-sending and receiving
// process. The Sender sends predefined messages over a named pipe,
// while the Receiver waits for incoming messages. The main function
// orchestrates these actions, showcasing different cases where a peer is
// present or absent for message exchange.

// PipePath defines the path for the named pipe in the system's temp directory.
var PipePath = filepath.Join(os.TempDir(), "test_pipe")
var wg sync.WaitGroup

// Sender struct encapsulates a named pipe used for sending messages.
type Sender struct {
	Pipe *namedpipes.NamedPipe
}

// Print formats and prints messages for the Sender.
func (s *Sender) Print(msg string, args ...any) {
	fmt.Printf(msg+"\n", args...)
}

// CheckPeer attempts to send a test message and verifies if a peer is connected within the given timeout.
func (s *Sender) CheckPeer(timeout float64) {
	s.Print("-- Start Sender")
	defer s.Print("-- Stop Sender")

	msg := "HELLO"
	s.Print("-- sending >> " + msg)
	// Attempt to send message with timeout to check for peer.
	if err := s.Pipe.Write([]byte(msg), timeout); err != nil {
		if errors.Is(err, namedpipes.ErrTimeout) {
			s.Print("-- TIMEOUT: no peer connected")
		} else {
			s.Print("-- FAILED: %v", err)
		}
	}

	time.Sleep(100 * time.Millisecond)
}

// SendMessages sends a sequence of messages over the pipe with the specified timeout.
func (s *Sender) SendMessages(timeout float64) {
	s.Print("-- Start Sender")
	defer s.Print("-- Stop Sender")

	for i := 1; i <= 5; i++ {
		msg := fmt.Sprintf("MESSAGE_%d", i)
		s.Print("-- sending >> " + msg)
		// Send each message and check for write errors.
		if err := s.Pipe.Write([]byte(msg), timeout); err != nil {
			s.Print("-- FAILED: %v", err)
		}
		time.Sleep(time.Millisecond * 500)
	}
}

// Receiver struct encapsulates a named pipe used for receiving messages.
type Receiver struct {
	Pipe *namedpipes.NamedPipe
}

// Print formats and prints messages for the Receiver.
func (r *Receiver) Print(msg string, args ...any) {
	fmt.Printf("                         "+msg+"\n", args...)
}

// WaitMessages reads messages from the pipe until an error or break signal is received.
func (r *Receiver) WaitMessages(timeout float64) {
	defer wg.Done()

	r.Print("-- Start Receiver")
	defer r.Print("-- Stop Receiver")

	for {
		// Attempt to read from pipe, handling different error cases.
		b, err := r.Pipe.Read(timeout)
		if err != nil {
			if err == namedpipes.ErrBreak {
				break
			} else {
				r.Print("-- %v", err)
			}
		} else {
			r.Print("-- received << %s", b)
		}
	}
}

func main() {
	fmt.Printf("\n**** starting ****\n")

	fmt.Printf("\nUsing Pipe: %s\n", PipePath)

	// Create named pipe with specified permissions, handling creation error.
	if err := namedpipes.Create(PipePath, 0o666); err != nil {
		fmt.Printf("Failed to create pipe: %v\n", err)
		return
	}
	defer namedpipes.Delete(PipePath)

	// Initialize Sender and Receiver with the created named pipe.
	sender := &Sender{
		Pipe: namedpipes.New(PipePath, nil),
	}
	receiver := &Receiver{
		Pipe: namedpipes.New(PipePath, nil),
	}

	// Check peer status without a connected peer.
	fmt.Printf("\n\n* checking with no peer:\n")
	sender.CheckPeer(2.0)

	// Check peer status with a connected peer.
	fmt.Printf("\n\n* checking with peer:\n")
	wg.Add(1)
	go receiver.WaitMessages(5.0)
	time.Sleep(10 * time.Millisecond)
	sender.CheckPeer(2.0)
	time.Sleep(100 * time.Millisecond)
	receiver.Pipe.Cancel()
	wg.Wait()

	// Send multiple messages with a connected receiver.
	fmt.Printf("\n\n* sending and receiving messages:\n")
	wg.Add(1)
	go receiver.WaitMessages(5.0)
	time.Sleep(10 * time.Millisecond)
	sender.SendMessages(2.0)
	time.Sleep(100 * time.Millisecond)
	receiver.Pipe.Cancel()
	wg.Wait()

	fmt.Printf("\n**** exit ****\n\n")
}
