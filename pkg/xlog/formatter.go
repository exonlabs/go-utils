package xlog

import (
	"fmt"
	"strings"
)

type Formatter struct {
	RecordFormat string
	TimeFormat   string
}

func NewFormatter(recFmt, tsFmt string) *Formatter {
	return &Formatter{recFmt, tsFmt}
}

func (f *Formatter) ParseRecord(r *Record) string {
	if r == nil {
		return ""
	} else {
		return strings.NewReplacer(
			"{time}", r.Time.Format(f.TimeFormat),
			"{level}", StringLevel(r.Level),
			"{source}", r.Source,
			"{message}", fmt.Sprintf(r.Message, r.MsgArgs...),
		).Replace(f.RecordFormat)
	}
}

// ///////////////////// creator functions

func NewStdFrmt() *Formatter {
	return &Formatter{
		RecordFormat: "{time} {level} {message}",
		TimeFormat:   "2006-01-02 15:04:05.000000",
	}
}

func NewStdFrmtWithSrc() *Formatter {
	return &Formatter{
		RecordFormat: "{time} {level} [{source}] {message}",
		TimeFormat:   "2006-01-02 15:04:05.000000",
	}
}

func NewCustomMsgFrmt(recFmt string) *Formatter {
	return &Formatter{
		RecordFormat: recFmt,
		TimeFormat:   "2006-01-02 15:04:05.000000",
	}
}

func NewCustomTimeFrmt(tsFmt string) *Formatter {
	return &Formatter{
		RecordFormat: "{time} {level} {message}",
		TimeFormat:   tsFmt,
	}
}
