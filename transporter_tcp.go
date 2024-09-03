package fins

import (
	"github.com/expgo/factory"
	"net"
	"time"
)

type TcpTransporter struct {
	baseTransporter
	conn *net.TCPConn
}

func newTcpTransport(addr string) *TcpTransporter {
	return factory.NewWithFunc[TcpTransporter](func() *TcpTransporter {
		return &TcpTransporter{baseTransporter: baseTransporter{addr: addr}}
	})
}

func (t *TcpTransporter) Open() error {
	if t.state == StateConnected {
		return nil
	}

	t.setState(StateConnecting)
	tcpAddr, err := net.ResolveTCPAddr("tcp", t.addr)
	if err != nil {
		t.L.Warnf("Resolve TCPAddr %s failed: %v", t.addr, err)
		t.setState(StateDisconnected)
		return err
	}

	t.conn, err = net.DialTCP("tcp", nil, tcpAddr)
	if err != nil {
		t.L.Warnf("DialTCP %s failed: %v", t.addr, err)
		t.setState(StateDisconnected)
		return err
	}

	t.setState(StateConnected)

	return nil
}

func (t *TcpTransporter) Close() error {
	defer func() {
		t.conn = nil
		t.setState(StateDisconnected)
	}()

	return t.conn.Close()
}

func (t *TcpTransporter) Write(data []byte) (n int, err error) {
	defer func() {
		if err != nil {
			t.setState(StateDisconnected)
		}
	}()

	err = t.conn.SetWriteDeadline(time.Now().Add(t.WriteTimeout))
	if err != nil {
		return 0, err
	}

	return t.conn.Write(data)
}

func (t *TcpTransporter) Read(buf []byte) (n int, err error) {
	defer func() {
		if err != nil {
			t.setState(StateDisconnected)
		}
	}()

	err = t.conn.SetReadDeadline(time.Now().Add(t.ReadTimeout))
	if err != nil {
		return 0, err
	}

	return t.conn.Read(buf)
}
