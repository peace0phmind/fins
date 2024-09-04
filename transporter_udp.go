package fins

import (
	"bytes"
	"encoding/binary"
	"github.com/expgo/factory"
	"net"
	"time"
)

type UdpTransporter struct {
	baseTransporter
	conn *net.UDPConn
}

func newUdpTransport(addr string) *UdpTransporter {
	return factory.NewWithFunc[UdpTransporter](func() *UdpTransporter {
		return &UdpTransporter{baseTransporter: baseTransporter{addr: addr}}
	})
}

func (t *UdpTransporter) Open() error {
	if t.state == StateConnected {
		return nil
	}

	t.setState(StateConnecting)
	serverAddr, err := net.ResolveUDPAddr("udp", t.addr)
	if err != nil {
		t.L.Warnf("Resolve UDPAddr %s failed: %v", t.addr, err)
		t.setState(StateDisconnected)
		return err
	}

	t.conn, err = net.DialUDP("udp", nil, serverAddr)
	if err != nil {
		t.L.Warnf("DialUDP %s failed: %v", t.addr, err)
		t.setState(StateDisconnected)
		return err
	}

	t.setState(StateConnected)

	return nil
}

func (t *UdpTransporter) Close() error {
	defer func() {
		t.conn = nil
		t.setState(StateDisconnected)
	}()

	return t.conn.Close()
}

func (t *UdpTransporter) Write(data []byte) (n int, err error) {
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

func (t *UdpTransporter) ReadHeader() (header *respFinsHeader, err error) {
	defer func() {
		if err != nil {
			t.setState(StateDisconnected)
		}
	}()

	headerBuf := make([]byte, respHeaderSize)
	_, err = t.ReadData(headerBuf)
	if err != nil {
		return nil, err
	}

	header = &respFinsHeader{}
	err = binary.Read(bytes.NewReader(headerBuf), binary.BigEndian, header)
	if err != nil {
		return nil, err
	}

	return header, nil
}

func (t *UdpTransporter) ReadData(buf []byte) (n int, err error) {
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
