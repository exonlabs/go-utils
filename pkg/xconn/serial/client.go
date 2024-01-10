package serial

import (
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/exonlabs/go-utils/pkg/xconn"
	"github.com/exonlabs/go-utils/pkg/xlog"
	"go.bug.st/serial"
)

type SerialClient struct {
	*BaseSerial
	portConn serial.Port
}

func NewSerialClient(port string, baudrate int, mode string,
	log *xlog.Logger) *SerialClient {
	return &SerialClient{
		BaseSerial: NewBaseSerial(port, baudrate, mode, log),
	}
}

func (conn *SerialClient) Type() string {
	p, sb := optsModeToPrim(conn.mode.Parity, conn.mode.StopBits)
	return fmt.Sprintf("<%v: %v %v:%v%v%v>",
		conn.GetType(), conn.Port, conn.mode.BaudRate,
		conn.mode.DataBits, p, sb)
}

func (conn *SerialClient) IsOpened() bool {
	if conn.portConn != nil {
		return true
	}
	return false
}

func (conn *SerialClient) Open() error {
	// close existing com if port value changed
	if conn.portConn != nil {
		if !conn.IsOpened() {
			conn.portConn.Close()
		}
	}

	var err error
	conn.portConn, err = serial.Open(conn.Port, &conn.mode)
	if err != nil {
		return err
	}

	conn.EvtKill.Clear()

	if conn.Log != nil {
		conn.Log.Info("OPEN -- %v", conn.Type())
	}

	return nil
}

func (conn *SerialClient) Close() error {
	if err := conn.portConn.Close(); err != nil {
		return err
	}

	conn.EvtKill.Set()

	conn.portConn = nil

	if conn.Log != nil {
		conn.Log.Info("CLOSE -- %v", conn.Type())
	}

	return nil
}

func (conn *SerialClient) Send(data []byte) error {
	if conn.portConn == nil {
		return xconn.ErrNotOpend
	}

	if len(data) == 0 {
		return errors.New("EMPTY_DATA")
	}

	conn.TxLog(data)

	_, err := conn.portConn.Write(data)
	if err != nil {
		return err
	}
	return nil
}

func (conn *SerialClient) Recv() ([]byte, error) {
	if conn.portConn == nil {
		return nil, xconn.ErrNotOpend
	}

	var res []byte
	var errStr string
	chunkRead := 1024

	for {
		buff := make([]byte, chunkRead)
		conn.portConn.SetReadTimeout(
			time.Millisecond * time.Duration(0.1*1000))
		n, err := conn.portConn.Read(buff)
		if err != nil {
			errStr = err.Error()
			break
		}

		if n > 0 {
			res = append(res, buff[:n]...)
		}

		if conn.EvtBreak.IsSet() {
			errStr = "BREAK_REQUEST"
			break
		}

		if conn.PollMaxSize != 0 && len(res) > conn.PollMaxSize {
			errStr = "MAX_DATA_LIMIT - reached max receive limit"
			break
		}

		if n < chunkRead && n == 0 {
			break
		}
	}

	if len(errStr) > 0 {
		// if errStr != io.EOF.Error() {
		return nil, errors.New(errStr)
		// }
	}

	if len(res) > 0 {
		conn.RxLog(res)
	}

	return res, nil
}

func (conn *SerialClient) RecvWait(timeout float64) ([]byte, error) {
	conn.EvtBreak.Clear()
	tlimit := time.Now().Add(time.Millisecond * time.Duration(timeout*1000))

	for {
		data, err := conn.Recv()
		if err != nil {
			return nil, err
		}

		if len(data) > 0 {
			return data, nil
		}

		if conn.EvtKill.IsSet() {
			return nil, xconn.ErrClosed
		}

		if timeout > 0 {
			if time.Now().After(tlimit) {
				return nil, xconn.ErrTimeout
			}
		}
	}
}

func (conn *SerialClient) Cancel() error {
	conn.EvtBreak.Set()
	return nil
}

// return databits, parity, stopbits
func strToOptsMode(mode string) (databits int, p serial.Parity, sb serial.StopBits) {
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
func optsModeToPrim(p serial.Parity, sb serial.StopBits) (string, float32) {
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
