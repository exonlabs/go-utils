package main

import (
	"bytes"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"runtime/debug"
	"syscall"

	"github.com/exonlabs/go-utils/pkg/unix/xpipe"
	"github.com/exonlabs/go-utils/pkg/xlog"
	"github.com/exonlabs/go-utils/pkg/xputil"
)

var (
	tmp_path     = filepath.Join(os.TempDir(), "SampleProcess")
	inpipe_file  = filepath.Join(tmp_path, "in.pipe")
	outpipe_file = filepath.Join(tmp_path, "out.pipe")
)

type SampleProcess struct {
	*xputil.BaseProcess
	counter int
}

func NewSampleProcess(log *xlog.Logger) *SampleProcess {
	pr := &SampleProcess{}
	pr.BaseProcess = xputil.NewBaseProcess(log, pr)
	return pr
}

func (pr *SampleProcess) Initialize() error {
	pr.Log.Info("initialized")
	return nil
}

func (pr *SampleProcess) Execute() error {
	pr.counter += 1
	pr.Log.Info("running: ... %d", pr.counter)
	pr.Sleep(1)
	return nil
}

func (pr *SampleProcess) Terminate() error {
	pr.Log.Info("terminated")
	return nil
}

func (pr *SampleProcess) HandleCommand(cmd string) string {
	pr.Log.Info("received command: %s", cmd)

	reply := "done"

	switch cmd {
	case "exit":
		pr.Stop()
	case "reset":
		pr.counter = 0
	default:
		reply = "invalid_command"
	}

	pr.Log.Info("reply command: %s", reply)
	return reply
}

func main() {
	logger := xlog.NewStdoutLogger("main")

	defer func() {
		if r := recover(); r != nil {
			stack := debug.Stack()
			indx := bytes.Index(stack, []byte("panic({"))
			logger.Panic("%s", r)
			logger.Trace1("\n----------\n%s----------", stack[indx:])
			os.Exit(1)
		} else {
			logger.Info("exit")
			os.Exit(0)
		}
	}()

	debug := flag.Int("x", 0, "set debug modes, (default: 0)")
	flag.Parse()

	switch {
	case *debug >= 5:
		logger.Level = xlog.TRACE4
	case *debug >= 4:
		logger.Level = xlog.TRACE3
	case *debug >= 3:
		logger.Level = xlog.TRACE2
	case *debug >= 2:
		logger.Level = xlog.TRACE1
	case *debug >= 1:
		logger.Level = xlog.DEBUG
	}

	logger.Info("**** starting ****")

	syscall.Umask(0)
	os.MkdirAll(tmp_path, 0o775)
	defer os.RemoveAll(tmp_path)

	logger.Info("Using Input Pipe: %s", inpipe_file)
	if err := xpipe.CreatePipe(inpipe_file, 0o666); err != nil {
		fmt.Println(err.Error())
		return
	}
	logger.Info("Using Output Pipe: %s", outpipe_file)
	if err := xpipe.CreatePipe(outpipe_file, 0o666); err != nil {
		fmt.Println(err.Error())
		return
	}

	pr := NewSampleProcess(logger)
	pr.ProcTitle = "SampleProcess"
	pr.InitManagement(
		xpipe.NewPipe(inpipe_file),
		xpipe.NewPipe(outpipe_file),
		pr.HandleCommand,
	)

	pr.Start()
}
