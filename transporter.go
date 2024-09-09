package fins

import (
	"github.com/expgo/log"
	"sync"
	"time"
)

/*
State

	@EnumConfig(marshal, noCase)
	@Enum {
		Unknown
		Connecting
		Connected
		Disconnected
		ConnectClosed
	}
*/
type State int

type Transporter interface {
	Open() error
	Close() error
	Write(header *finsHeader, data []byte) (int, error)
	ReadHeader() (*respFinsHeader, error)
	ReadData(buf []byte) (int, error)
	State() State
	setState(state State, err error)
	SetStateChangeCallback(callback func(oldState, newState State))
}

type baseTransporter struct {
	log.InnerLog
	ReadTimeout          time.Duration `value:"3s"`
	WriteTimeout         time.Duration `value:"3s"`
	ReconnectionInterval time.Duration `value:"10s"`
	addr                 string

	reconnectTimer *time.Timer
	state          State       `value:"unknown"`
	self           Transporter `wire:"self"`
	callback       func(oldState, newState State)
	stateLock      sync.Mutex
}

func (t *baseTransporter) State() State {
	return t.state
}
func (t *baseTransporter) SetStateChangeCallback(callback func(oldState, newState State)) {
	t.callback = callback
}

func (t *baseTransporter) setState(state State, err error) {
	t.stateLock.Lock()
	defer t.stateLock.Unlock()

	if state == StateDisconnected {
		t.startReconnectTimer()
	}

	if t.callback != nil {
		t.callback(t.state, state)
	}

	t.L.Infof("%s state change, old state: %s, new state: %s, err: %v", t.addr, t.state, state, err)

	t.state = state
}

func (t *baseTransporter) startReconnectTimer() {
	if t.ReconnectionInterval <= 0 {
		return
	}

	if t.reconnectTimer == nil {
		t.reconnectTimer = time.AfterFunc(t.ReconnectionInterval, t.reconnect)
	} else {
		t.reconnectTimer.Reset(t.ReconnectionInterval)
	}
}

func (t *baseTransporter) reconnect() {
	if t.reconnectTimer != nil {
		t.reconnectTimer.Stop()
		t.reconnectTimer = nil
	}

	_ = t.self.Close()
	_ = t.self.Open()
}
