package xlog

import "os"

type Handler interface {
	SetFormatter(Formatter)
	HandleRecord(Record) error
}

// /////////////////////

type BaseHandler struct {
	Frmt Formatter
}

func (h *BaseHandler) SetFormatter(f Formatter) {
	h.Frmt = f
}

func NewBaseHandler() *BaseHandler {
	return &BaseHandler{
		Frmt: GetDefaultFormatter(),
	}
}

// ///////////////////// Write log messages to Stdout

type StdoutHandler struct {
	*BaseHandler
}

func NewStdoutHandler() *StdoutHandler {
	return &StdoutHandler{
		BaseHandler: NewBaseHandler(),
	}
}

func (h *StdoutHandler) HandleRecord(r Record) error {
	_, err := os.Stdout.WriteString(h.Frmt.ParseRecord(r) + "\n")
	return err
}

// ///////////////////// Write log messages to file

type FileHandler struct {
	*BaseHandler
	FilePath string
}

func NewFileHandler(path string) *FileHandler {
	return &FileHandler{
		BaseHandler: NewBaseHandler(),
		FilePath:    path,
	}
}

func (h *FileHandler) HandleRecord(r Record) error {
	fh, err := os.OpenFile(
		h.FilePath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0664)
	if err != nil {
		return err
	}
	defer fh.Close()
	_, err = fh.WriteString(h.Frmt.ParseRecord(r) + "\n")
	return err
}
