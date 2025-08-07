// Copyright (c) 2024 ExonLabs, All rights reserved.
// Use of this source code is governed by a BSD 3-Clause
// license that can be found in the LICENSE file.

package proc

import (
	"bytes"
	"errors"
	"fmt"
	"runtime/debug"

	"github.com/exonlabs/go-utils/pkg/comm"
	"github.com/exonlabs/go-utils/pkg/logging"
	"github.com/exonlabs/go-utils/pkg/tasklet"
)

var (
	ErrCommListener = errors.New("invalid comm listener")
	ErrDataHandler  = errors.New("invalid IPC data handler")
)

type IpcDataHandler func([]byte) ([]byte, error)

type IpcListener struct {
	*tasklet.AsyncTasklet

	commListener comm.Listener
	dataHandler  IpcDataHandler
}

func NewIpcListener(log *logging.Logger, l comm.Listener, h IpcDataHandler) (*IpcListener, error) {
	if l == nil {
		return nil, ErrCommListener
	}
	if h == nil {
		return nil, ErrDataHandler
	}

	ipc := &IpcListener{
		commListener: l,
		dataHandler:  h,
	}
	ipc.commListener.SetConnHandler(ipc.handleConnection)
	ipc.AsyncTasklet = tasklet.NewAsyncTasklet(log, ipc)
	return ipc, nil
}

func (ipc *IpcListener) Initialize() error {
	return nil
}

func (ipc *IpcListener) Execute() error {
	return ipc.commListener.Start()
}

func (ipc *IpcListener) Terminate() error {
	ipc.commListener.Stop()
	return nil
}

func (ipc *IpcListener) handleConnection(conn comm.Connection) {
	defer func() {
		if r := recover(); r != nil {
			stack := debug.Stack()
			indx := bytes.Index(stack, []byte("panic({"))
			if ipc.Log != nil {
				ipc.Log.Panic("%v\n----------\n%s----------", r, stack[indx:])
			} else {
				fmt.Printf("%v\n----------\n%s----------\n", r, stack[indx:])
			}
		}
	}()

	for conn.IsOpened() {
		data, addr, err := conn.RecvFrom(-1)
		if err != nil {
			if !conn.IsOpened() || errors.Is(err, comm.ErrClosed) {
				return
			}
			if ipc.Log != nil {
				ipc.Log.Error("error receiving:", err.Error())
			} else {
				fmt.Println("error receiving:", err.Error())
			}
			continue
		}

		// handle received data
		reply, err := ipc.dataHandler(data)
		if err != nil {
			if ipc.Log != nil {
				ipc.Log.Error("error handling received data:", err.Error())
			} else {
				fmt.Println("error handling received data:", err.Error())
			}
		}

		// sending reply if any
		if len(reply) > 0 {
			if err = conn.SendTo(reply, addr, -1); err != nil {
				if ipc.Log != nil {
					ipc.Log.Error("error sending reply:", err.Error())
				} else {
					fmt.Println("error sending reply:", err.Error())
				}
			}
		}
	}
}
