package xcomm

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/exonlabs/go-utils/pkg/xlog"
	"go.bug.st/serial"
)

const (
	defaultXONXOFF = bool(false)
	defaultRTSCTS  = bool(false)
	defaultDSRDTR  = bool(false)
)

// Serial Connection URI
//
// format:  serial@<port>:<baud>:<mode>[:<opts>]
//   <port>  com port name
//   <baud>  baudrate
//   <mode>  bytesize, parity and stopbits
//           {8|7}{N|E|O|M|S}{1|2}
//   <opts>  {rtscts|dsrdtr|xonxoff}
//
// example:
//   - serial@/dev/ttyS0:115200:8N1          (linux)
//   - serial@/dev/ttyS0:115200:8N1:rtscts   (linux)
//   - serial@COM1:115200:8N1         (win)

// parse and validate uri
func parseSerialURI(uri string) (string, serial.Mode, error) {
	p := strings.SplitN(uri, "@", 2)
	if len(p) < 2 || p[0] != "serial" {
		return "", serial.Mode{}, ErrInvalidUri
	}

	p = strings.Split(p[1], ":")
	if len(p) < 3 || len(p[2]) != 3 {
		return "", serial.Mode{}, ErrInvalidUri
	}

	var err error

	mode := serial.Mode{}
	mode.BaudRate, err = strconv.Atoi(p[1])
	if err != nil {
		return "", serial.Mode{}, ErrInvalidUri
	}
	mode.DataBits, err = strconv.Atoi(string(p[2][0]))
	if err != nil {
		return "", serial.Mode{}, ErrInvalidUri
	}
	switch string(p[2][1]) {
	case "n", "N":
		mode.Parity = serial.NoParity
	case "o", "O":
		mode.Parity = serial.OddParity
	case "e", "E":
		mode.Parity = serial.EvenParity
	case "m", "M":
		mode.Parity = serial.MarkParity
	case "s", "S":
		mode.Parity = serial.SpaceParity
	default:
		return "", serial.Mode{}, ErrInvalidUri
	}
	switch string(p[2][2]) {
	case "1":
		mode.StopBits = serial.OneStopBit
	case "2":
		mode.StopBits = serial.TwoStopBits
	default:
		return "", serial.Mode{}, ErrInvalidUri
	}
	// mode.InitialStatusBits = &serial.ModemOutputBits{
	// 	RTS: false,
	// 	DTR: false,
	// }

	return p[0], mode, nil

}

// //////////////////////////////////////////////////

// Serial Connection
type SerialConnection struct {
	*BaseConnection
	port string
	mode serial.Mode

	// low level serial port handler
	com serial.Port

	// parent server handler
	parent *SerialListener
}

func NewSerialConnection(uri string, log *xlog.Logger) (*SerialConnection, error) {
	sc := &SerialConnection{
		BaseConnection: NewBaseConnection(uri, log),
	}
	var err error
	sc.port, sc.mode, err = parseSerialURI(uri)
	if err != nil {
		return nil, err
	}
	return sc, nil
}

func (sc *SerialConnection) Parent() Listener {
	return sc.parent
}

func (sc *SerialConnection) IsOpened() bool {
	return bool(sc.com != nil)
}

func (sc *SerialConnection) Open() error {
	sc.Close()

	sc.evtBreak.Clear()
	sc.evtKill.Clear()

	sc.log("OPEN -- %s", sc.uri)

	var err error
	sc.com, err = serial.Open(sc.port, &sc.mode)
	if err != nil {
		return fmt.Errorf("%w, %s", ErrOpen, err.Error())
	}

	return nil
}

// Closes the serial connection
func (sc *SerialConnection) Close() {
	sc.evtKill.Set()
	if sc.com != nil {
		sc.com.Close()
		sc.log("CLOSE -- %s", sc.uri)
	}
	sc.com = nil
}

// Sends data over the serial connection
func (sc *SerialConnection) Send(data []byte) error {
	if data == nil || len(data) == 0 {
		return fmt.Errorf("%w, empty data", ErrError)
	}
	if sc.com == nil {
		return ErrNotOpend
	}

	sc.txLog(data)
	_, err := sc.com.Write(data)
	sc.com.Drain()
	if err != nil {
		if errIsClosed(err) {
			sc.Close()
		}
		return fmt.Errorf("%w, %s", ErrWrite, err.Error())
	}

	return nil
}

// Recv data from the socket connection
func (sc *SerialConnection) Recv() ([]byte, error) {
	if sc.com == nil {
		return nil, ErrNotOpend
	}

	data := []byte(nil)
	for {
		b := make([]byte, sc.PollChunkSize)

		sc.com.SetReadTimeout(
			time.Duration(sc.PollInterval * 1000000000))

		n, err := sc.com.Read(b)
		if err != nil {
			if errIsClosed(err) {
				sc.rxLog(data)
				sc.Close()
				return nil, ErrClosed
			}
			return nil, fmt.Errorf("%w, %s", ErrRead, err.Error())
		}
		if n > 0 {
			data = append(data, b[:n]...)
		} else {
			break
		}

		if sc.PollMaxSize > 0 && len(data) > sc.PollMaxSize {
			break
		}

		if sc.evtKill.IsSet() {
			sc.rxLog(data)
			return nil, ErrClosed
		}
		if sc.evtBreak.IsSet() {
			sc.rxLog(data)
			return nil, ErrBreak
		}
	}
	sc.rxLog(data)
	return data, nil
}

// Receives data with a specified timeout from the socket connection
func (sc *SerialConnection) RecvWait(timeout float64) ([]byte, error) {
	sc.evtBreak.Clear()
	tbreak := float64(time.Now().Unix()) + timeout
	for {
		data, err := sc.Recv()
		if err != nil {
			return nil, err
		} else if data != nil && len(data) > 0 {
			return data, nil
		}
		if sc.evtKill.IsSet() {
			return nil, ErrClosed
		}
		if sc.evtBreak.IsSet() {
			return nil, ErrBreak
		}
		if timeout > 0 && float64(time.Now().Unix()) >= tbreak {
			return nil, ErrTimeout
		}
	}
}

// cancel blocking operations
func (sc *SerialConnection) Cancel() {
	sc.evtBreak.Set()
}

// //////////////////////////////////////////////////

// Serial Listener
type SerialListener struct {
	*SerialConnection

	// callback connection handler function
	connHandler func(Connection)
}

func NewSerialListener(uri string, log *xlog.Logger) (*SerialListener, error) {
	sc, err := NewSerialConnection(uri, log)
	if err != nil {
		return nil, err
	}
	return &SerialListener{
		SerialConnection: sc,
	}, nil
}

func (sl *SerialListener) IsActive() bool {
	return sl.IsOpened()
}

func (sl *SerialListener) Start() error {
	if sl.connHandler == nil {
		return fmt.Errorf("%w, invalid connection handler", ErrOpen)
	}

	if err := sl.Open(); err != nil {
		return err
	}

	sl.run()
	return nil
}

func (sl *SerialListener) run() {
	for !sl.evtKill.IsSet() {
		sl.connHandler(sl)
		sl.Sleep(1)
	}
}

func (sl *SerialListener) Stop() {
	sl.Close()
}

func (sl *SerialListener) SetHandler(f func(Connection)) {
	sl.connHandler = f
}