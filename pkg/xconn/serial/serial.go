package serial

import (
	"strconv"

	"github.com/exonlabs/go-utils/pkg/xconn"

	"github.com/exonlabs/go-utils/pkg/xlog"
	"go.bug.st/serial"
)

type BaseSerial struct {
	*xconn.BaseConnection
	Port string
	mode serial.Mode
}

func NewBaseSerial(port string, baudrate int, mode string, log *xlog.Logger) *BaseSerial {
	conn := &BaseSerial{
		BaseConnection: xconn.NewBaseConnection("", log),
		Port:           port,
	}
	conn.mode.BaudRate = baudrate
	conn.mode.DataBits,
		conn.mode.Parity,
		conn.mode.StopBits = strToOptsMode(mode)
	conn.mode.InitialStatusBits = &serial.ModemOutputBits{
		RTS: false,
		DTR: false,
	}

	return conn
}

func (*BaseSerial) GetType() string {
	return "SERIAL"
}

// return databits, parity, stopbits
func strToOptsMode2(mode string) (databits int, p serial.Parity, sb serial.StopBits) {
	databits, _ = strconv.Atoi(string(mode[0]))

	// Parity
	switch string(mode[1]) {
	case "O":
		p = 1
	case "E":
		p = 2
	case "M":
		p = 3
	case "S":
		p = 4
	default:
		p = 0
	}

	// StopBits
	stopBits, _ := strconv.ParseFloat(string(mode[2:]), 64)
	switch stopBits {
	case 1.5:
		sb = serial.OnePointFiveStopBits
	case 2:
		sb = serial.TwoStopBits
	default:
		sb = serial.OneStopBit
	}

	return
}

// convert pkg data types to primitive data types
// parity string, stopbits float32
func optsModeToPrim2(p serial.Parity, sb serial.StopBits) (string, float32) {
	var (
		parity   string
		stopbits float32
	)

	switch p {
	case 1:
		parity = "O"
	case 2:
		parity = "E"
	case 3:
		parity = "M"
	case 4:
		parity = "S"
	default:
		parity = "N"
	}

	switch sb {
	case serial.OnePointFiveStopBits:
		stopbits = 1.5
	case serial.TwoStopBits:
		stopbits = 2
	default:
		stopbits = 1
	}

	return parity, stopbits
}
