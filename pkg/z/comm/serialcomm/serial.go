package xcomm

import (
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	"go.bug.st/serial"
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
func parse_serial_uri(uri string) (string, serial.Mode, error) {
	var err error

	p := strings.SplitN(uri, "@", 2)
	if len(p) < 2 || strings.ToLower(p[0]) != "serial" {
		return "", serial.Mode{}, ErrUri
	}
	p = strings.Split(p[1], ":")
	if len(p) < 3 || len(p[2]) != 3 {
		return "", serial.Mode{}, ErrUri
	}

	mode := serial.Mode{}
	mode.BaudRate, err = strconv.Atoi(p[1])
	if err != nil {
		return "", serial.Mode{}, ErrUri
	}
	mode.DataBits, err = strconv.Atoi(string(p[2][0]))
	if err != nil {
		return "", serial.Mode{}, ErrUri
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
		return "", serial.Mode{}, ErrUri
	}
	switch string(p[2][2]) {
	case "1":
		mode.StopBits = serial.OneStopBit
	case "2":
		mode.StopBits = serial.TwoStopBits
	default:
		return "", serial.Mode{}, ErrUri
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
	*baseConnection

	port string
	mode serial.Mode

	// low level serial port handler
	com serial.Port
	// parent server handler
	parent *SerialListener
}

func NewSerialConnection(
	uri string, log *Logger, opts Options) (*SerialConnection, error) {
	var err error
	sc := &SerialConnection{
		baseConnection: new_base_connection(uri, log, opts),
	}
	sc.port, sc.mode, err = parse_serial_uri(uri)
	if err != nil {
		return nil, fmt.Errorf("%w%s", ErrError, err)
	}
	// set max polling relative to baudrate. set to more than 10 times
	// actual byte duration (10 bits)
	sc.PollInterval = math.Max(sc.PollInterval,
		math.Ceil(100000000/float64(sc.mode.BaudRate))/1000000)
	return sc, nil
}

func (sc *SerialConnection) Parent() Listener {
	return sc.parent
}

func (sc *SerialConnection) PortHandler() serial.Port {
	return sc.com
}

func (sc *SerialConnection) IsOpened() bool {
	return !(sc.evtKill.IsSet() || sc.com == nil)
}

func (sc *SerialConnection) Open() error {
	if !sc.op_mux.TryLock() {
		return nil
	}
	defer sc.op_mux.Unlock()

	sc.evtBreak.Clear()
	sc.evtKill.Clear()

	if sc.uriLogging {
		sc.comm_log("OPENED")
	} else {
		sc.comm_log("OPEN -- %s", sc.uri)
	}

	var err error

	sc.com, err = serial.Open(sc.port, &sc.mode)
	if err != nil {
		return fmt.Errorf("%w, %s", ErrConnection, err)
	}

	return nil
}

// Closes the serial connection
func (sc *SerialConnection) Close() {
	sc.op_mux.Lock()
	defer sc.op_mux.Unlock()

	sc.evtKill.Set()
	if sc.com != nil {
		sc.com.Close()
		sc.com = nil
		if sc.uriLogging {
			sc.comm_log("CLOSED")
		} else {
			sc.comm_log("CLOSE -- %s", sc.uri)
		}
	}
}

// cancel blocking operations
func (sc *SerialConnection) Cancel() {
	sc.evtBreak.Set()
	if sc.ctxCancel != nil {
		sc.ctxCancel()
		sc.ctxCancel = nil
	}
}

// Sends data over the serial connection
func (sc *SerialConnection) Send(data []byte) error {
	if len(data) == 0 {
		return fmt.Errorf("%w, empty data", ErrError)
	}
	if !sc.IsOpened() {
		return ErrClosed
	}
	sc.tx_Log(data)
	_, err := sc.com.Write(data)
	sc.com.Drain()
	if err != nil {
		if errIsClosed(err) {
			sc.comm_log("PORT_CLOSED - %s", err.Error())
			sc.Close()
			return ErrClosed
		}
		return fmt.Errorf("%w, %s", ErrWrite, err)
	}
	return nil
}

// Recv data from the serial connection
func (sc *SerialConnection) Recv() ([]byte, error) {
	if !sc.IsOpened() {
		return nil, ErrClosed
	}

	data := []byte(nil)
	td := time.Duration(sc.PollInterval * 1000000000)
	if td > 0 {
		sc.com.SetReadTimeout(td)
	}

	for {
		b := make([]byte, sc.PollChunkSize)
		n, err := sc.com.Read(b)
		if err != nil {
			if errIsClosed(err) {
				sc.rx_Log(data)
				sc.comm_log("PORT_CLOSED - %s", err.Error())
				sc.Close()
				return nil, ErrClosed
			}
			return nil, fmt.Errorf("%w, %s", ErrRead, err)
		}
		if n > 0 {
			data = append(data, b[:n]...)
			if !sc.PollWaitNullRead && n < sc.PollChunkSize {
				break
			}
		} else {
			break
		}

		if sc.PollMaxSize > 0 && len(data) > sc.PollMaxSize {
			break
		}

		if sc.evtKill.IsSet() || (sc.parent != nil && !sc.parent.IsActive()) {
			sc.rx_Log(data)
			return nil, ErrClosed
		}
		if sc.evtBreak.IsSet() {
			sc.rx_Log(data)
			return nil, ErrBreak
		}
	}
	sc.rx_Log(data)
	return data, nil
}

// Receives data with a specified timeout from the serial connection
func (sc *SerialConnection) RecvWait(timeout float64) ([]byte, error) {
	sc.evtBreak.Clear()
	tbreak := float64(time.Now().Unix()) + timeout
	for {
		data, err := sc.Recv()
		if err != nil {
			return nil, err
		} else if len(data) > 0 {
			return data, nil
		}
		if sc.evtKill.IsSet() || (sc.parent != nil && !sc.parent.IsActive()) {
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

// //////////////////////////////////////////////////

// Serial Listener
type SerialListener struct {
	*SerialConnection

	// callback connection handler function
	connHandler func(Connection)
}

func NewSerialListener(
	uri string, log *Logger, opts Options) (*SerialListener, error) {
	sl := &SerialListener{}
	sc, err := NewSerialConnection(uri, log, opts)
	if err != nil {
		return nil, err
	}
	sc.parent = sl
	sl.SerialConnection = sc
	return sl, nil
}

func (sl *SerialListener) PortHandler() serial.Port {
	return sl.com
}

func (sl *SerialListener) SetConnHandler(h func(Connection)) {
	sl.connHandler = h
}

func (sl *SerialListener) IsActive() bool {
	return sl.IsOpened()
}

// no close action in listener mode, nust use stop method
func (sl *SerialListener) Close() {}

func (sl *SerialListener) Start() error {
	if sl.connHandler == nil {
		return fmt.Errorf("%wconnection handler not set", ErrError)
	}

	if err := sl.Open(); err != nil {
		return err
	}

	// run forever
	for sl.IsActive() {
		sl.connHandler(sl)
	}

	sl.SerialConnection.Close()
	return nil
}

func (sl *SerialListener) Stop() {
	sl.SerialConnection.Close()
}
