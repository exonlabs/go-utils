package xpipe

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"syscall"
	"time"

	"github.com/exonlabs/go-utils/pkg/sync/xevent"
	"github.com/exonlabs/go-utils/pkg/types"
	"golang.org/x/sys/unix"
)

const (
	DEFAULT_POLLINTERVAL  = float64(0.1)
	DEFAULT_POLLCHUNKSIZE = int(1024)
	DEFAULT_RWTIMEOUT     = float64(0)
)

var (
	ErrError   = errors.New("")
	ErrOpen    = fmt.Errorf("%wopen pipe failed", ErrError)
	ErrBreak   = fmt.Errorf("%woperation break", ErrError)
	ErrTimeout = fmt.Errorf("%woperation timeout", ErrError)
	ErrRead    = fmt.Errorf("%wread pipe failed", ErrError)
	ErrWrite   = fmt.Errorf("%wwrite pipe failed", ErrError)
)

type Options = types.NDict

type Pipe struct {
	fd       *os.File
	filePath string
	evtBreak *xevent.Event

	// read/write params
	PollInterval  float64
	PollChunkSize int
	ReadTimeout   float64
	WriteTimeout  float64
}

func NewPipe(path string, opts Options) *Pipe {
	return &Pipe{
		filePath:      filepath.Clean(path),
		evtBreak:      xevent.NewEvent(),
		PollInterval:  opts.GetFloat64("poll_interval", DEFAULT_POLLINTERVAL),
		PollChunkSize: opts.GetInt("poll_chunksize", DEFAULT_POLLCHUNKSIZE),
		ReadTimeout:   opts.GetFloat64("read_timeout", DEFAULT_RWTIMEOUT),
		WriteTimeout:  opts.GetFloat64("write_timeout", DEFAULT_RWTIMEOUT),
	}
}

// return pipe file path
func (p *Pipe) Path() string {
	return p.filePath
}

// open file connection handler on pipe
func (p *Pipe) open(mode int) error {
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
func (p *Pipe) open_read() error {
	return p.open(os.O_RDONLY | unix.O_NONBLOCK)
}

// open pipe for writing in non-blocking mode
func (p *Pipe) open_write() error {
	return p.open(os.O_WRONLY | unix.O_NONBLOCK)
}

// close file connection handler on pipe
func (p *Pipe) close() {
	if p.fd != nil {
		p.fd.Close()
	}
	p.fd = nil
}

// cancel active read/write operation on pipe
func (p *Pipe) Cancel() {
	p.evtBreak.Set()
}

// non-blocking read from pipe
func (p *Pipe) Read() ([]byte, error) {
	if p.fd == nil {
		if err := p.open_read(); err == nil {
			return nil, err
		}
		defer p.close()
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
		defer p.close()
	}
	if timeout < 0 {
		timeout = p.ReadTimeout
	}
	p.evtBreak.Clear()
	tbreak := float64(time.Now().Unix()) + timeout
	for {
		if err := p.open_read(); err == nil {
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

// non-blocking write to pipe
func (p *Pipe) Write(data []byte) error {
	if p.fd == nil {
		if err := p.open_write(); err == nil {
			return err
		}
		defer p.close()
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
		defer p.close()
	}
	if timeout < 0 {
		timeout = p.WriteTimeout
	}
	p.evtBreak.Clear()
	tbreak := float64(time.Now().Unix()) + timeout
	for {
		if err := p.open_write(); err == nil {
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

///////////////////////////////////// utils functions

// create new pipe on system, if not exist
func CreatePipe(path string, perm uint32) error {
	path = filepath.Clean(path)
	if path == string(filepath.Separator) || path == filepath.Dir(path) {
		return fmt.Errorf("%winvalid pipe path", ErrError)
	}
	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		if err := syscall.Mkfifo(path, perm); err != nil {
			return fmt.Errorf("%w%s", ErrError, err.Error())
		}
	} else if err != nil {
		return fmt.Errorf("%w%s", ErrError, err.Error())
	}
	return nil
}

// delete pipe file from system
func DeletePipe(path string) error {
	path = filepath.Clean(path)
	if path == string(filepath.Separator) || path == filepath.Dir(path) {
		return fmt.Errorf("%winvalid pipe path", ErrError)
	}
	err := os.Remove(path)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("%w%s", ErrError, err.Error())
	}
	return nil
}
