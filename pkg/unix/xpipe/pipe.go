package xpipe

import (
	"errors"
	"fmt"
	"io"
	"os"
	"syscall"
	"time"

	"github.com/exonlabs/go-utils/pkg/sync/xevent"
	"golang.org/x/sys/unix"
)

var (
	ErrError   = errors.New("")
	ErrOpen    = fmt.Errorf("%wopen pipe failed", ErrError)
	ErrClosed  = fmt.Errorf("%wpipe closed", ErrError)
	ErrBreak   = fmt.Errorf("%woperation break", ErrError)
	ErrTimeout = fmt.Errorf("%woperation timeout", ErrError)
	ErrRead    = fmt.Errorf("%wread pipe failed", ErrError)
	ErrWrite   = fmt.Errorf("%wwrite pipe failed", ErrError)
)

const (
	defaultPollInterval  = float64(0.1)
	defaultPollChunkSize = int(4096)
	defaultPollMaxSize   = int(0)
)

type Pipe struct {
	fd       *os.File
	filePath string
	evtBreak *xevent.Event

	// read/write polling params
	PollInterval  float64
	PollChunkSize int
	PollMaxSize   int
}

func NewPipe(path string) *Pipe {
	return &Pipe{
		filePath:      path,
		evtBreak:      xevent.NewEvent(),
		PollInterval:  defaultPollInterval,
		PollChunkSize: defaultPollChunkSize,
		PollMaxSize:   defaultPollMaxSize,
	}
}

// create new pipe dev-file on system, removes previous file if exist
func (p *Pipe) Create(perm uint32) error {
	os.Remove(p.filePath)
	if err := syscall.Mkfifo(p.filePath, perm); err != nil {
		return fmt.Errorf("%w%s", ErrError, err.Error())
	}
	return nil
}

// removes pipe dev-file from system
func (p *Pipe) Delete() error {
	err := os.Remove(p.filePath)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("%w%s", ErrError, err.Error())
	}
	return nil
}

// open file connection handler on pipe
func (p *Pipe) Open(mode int) error {
	if p.fd == nil {
		var err error
		p.fd, err = os.OpenFile(p.filePath, mode, os.ModeNamedPipe)
		if err != nil {
			return fmt.Errorf("%w, %s", ErrOpen, err.Error())
		}
	}
	return nil
}

// open pipe for reading in non-blocking mode
func (p *Pipe) OpenRead() error {
	return p.Open(os.O_RDONLY | unix.O_NONBLOCK)
}

// open pipe for writing in non-blocking mode
func (p *Pipe) OpenWrite() error {
	return p.Open(os.O_WRONLY | unix.O_NONBLOCK)
}

// close file connection handler on pipe
func (p *Pipe) Close() {
	if p.fd != nil {
		p.fd.Close()
		p.fd = nil
	}
}

// cancel active read/write operation on pipe
func (p *Pipe) Cancel() {
	p.evtBreak.Set()
}

// read from pipe
func (p *Pipe) Read() ([]byte, error) {
	if p.fd == nil {
		return nil, ErrClosed
	}
	data := []byte(nil)
	for {
		b := make([]byte, p.PollChunkSize)
		n, err := p.fd.Read(b)
		if err != nil {
			if err == io.EOF && data != nil {
				return data, nil
			}
			return nil, fmt.Errorf("%w, %s", ErrRead, err.Error())
		}
		if n > 0 {
			data = append(data, b[0:n]...)
		}
		if data != nil && (n == 0 || n < p.PollChunkSize) {
			return data, nil
		}
	}
}

// read from pipe, until data is received or timeout.
// use timeout=0 to wait forever (blocking mode)
func (p *Pipe) ReadWait(timeout float64) ([]byte, error) {
	if p.fd == nil {
		defer p.Close()
	}
	p.evtBreak.Clear()
	tbreak := float64(time.Now().Unix()) + timeout
	for {
		if err := p.OpenRead(); err == nil {
			if data, err := p.Read(); err == nil {
				return data, nil
			}
		}
		if !p.evtBreak.Wait(p.PollInterval) {
			return nil, ErrBreak
		}
		if timeout > 0 && float64(time.Now().Unix()) >= tbreak {
			return nil, ErrTimeout
		}
	}
}

// write to pipe
func (p *Pipe) Write(data []byte) error {
	if p.fd == nil {
		return ErrClosed
	}
	if _, err := p.fd.Write(data); err != nil {
		return fmt.Errorf("%w, %s", ErrWrite, err.Error())
	}
	return nil
}

// write to pipe, wait until peer is connected
// use timeout=0 to wait forever (blocking mode)
func (p *Pipe) WriteWait(data []byte, timeout float64) error {
	if p.fd == nil {
		defer p.Close()
	}
	p.evtBreak.Clear()
	tbreak := float64(time.Now().Unix()) + timeout
	for {
		if err := p.OpenWrite(); err == nil {
			return p.Write(data)
		}
		if !p.evtBreak.Wait(p.PollInterval) {
			return ErrBreak
		}
		if timeout > 0 && float64(time.Now().Unix()) >= tbreak {
			return ErrTimeout
		}
	}
}
