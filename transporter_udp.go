package fins

import (
	"bytes"
	"encoding/binary"
	"errors"
	"github.com/expgo/factory"
	"net"
	"time"
)

type UdpTransporter struct {
	baseTransporter
	conn *net.UDPConn
	da1  byte
	sa1  byte
}

func newUdpTransport(addr string) *UdpTransporter {
	return factory.NewWithFunc[UdpTransporter](func() *UdpTransporter {
		return &UdpTransporter{baseTransporter: baseTransporter{addr: addr},
			da1: 0xe8,
			sa1: 0x38,
		}
	})
}

func (t *UdpTransporter) Open() error {
	if t.state == StateConnected {
		return nil
	}

	t.setState(StateConnecting, nil)
	serverAddr, err := net.ResolveUDPAddr("udp", t.addr)
	if err != nil {
		t.L.Warnf("Resolve UDPAddr %s failed: %v", t.addr, err)
		t.setState(StateDisconnected, err)
		return err
	}

	t.conn, err = net.DialUDP("udp", nil, serverAddr)
	if err != nil {
		t.L.Warnf("DialUDP %s failed: %v", t.addr, err)
		t.setState(StateDisconnected, err)
		return err
	}

	t.setState(StateConnected, nil)

	return nil
}

func (t *UdpTransporter) Close() (err error) {
	defer func() {
		t.conn = nil
		t.setState(StateConnectClosed, err)
	}()

	return t.conn.Close()
}

func (t *UdpTransporter) Write(header *finsHeader, data []byte) (n int, err error) {
	if t.conn == nil || t.state == StateDisconnected {
		return 0, errors.New("udp transporter not connected")
	}

	defer func() {
		if err != nil {
			t.setState(StateDisconnected, err)
		}
	}()

	header.DA1 = t.da1
	header.SA1 = t.sa1

	buf := &bytes.Buffer{}

	err = binary.Write(buf, binary.BigEndian, header)
	if err != nil {
		return 0, err
	}

	if len(data) > 0 {
		_, err = buf.Write(data)
		if err != nil {
			return 0, err
		}
	}

	err = t.conn.SetWriteDeadline(time.Now().Add(t.WriteTimeout))
	if err != nil {
		return 0, err
	}

	return t.conn.Write(buf.Bytes())
}

func (t *UdpTransporter) ReadHeader() (header *respFinsHeader, err error) {
	defer func() {
		if err != nil {
			t.setState(StateDisconnected, err)
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
			t.setState(StateDisconnected, err)
		}
	}()

	err = t.conn.SetReadDeadline(time.Now().Add(t.ReadTimeout))
	if err != nil {
		return 0, err
	}

	return t.conn.Read(buf)
}
