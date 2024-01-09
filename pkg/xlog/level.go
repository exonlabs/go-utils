package xlog

type Level int

// logging levels
const (
	TRACE4 Level = -50
	TRACE3 Level = -40
	TRACE2 Level = -30
	TRACE1 Level = -20
	DEBUG  Level = -10
	INFO   Level = 0
	WARN   Level = 10
	ERROR  Level = 20
	FATAL  Level = 30
	PANIC  Level = 40
)

// returns text representation for log level
func StringLevel(l Level) string {
	switch {
	case l >= PANIC:
		return "PANIC"
	case l >= FATAL:
		return "FATAL"
	case l >= ERROR:
		return "ERROR"
	case l >= WARN:
		return "WARN "
	case l >= INFO:
		return "INFO "
	case l >= DEBUG:
		return "DEBUG"
	}
	return "TRACE"
}
