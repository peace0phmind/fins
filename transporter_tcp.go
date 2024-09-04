package fins

import (
	"bytes"
	"encoding/binary"
	"errors"
	"github.com/expgo/factory"
	"net"
	"time"
)

/*
TcpCommand

	@Enum {
		FrameSend = 0x2
	}
*/
type TcpCommand int

type tcpFinsHeader struct {
	Magic     [4]byte
	Length    uint32
	Command   TcpCommand
	ErrorCode uint32
}

func newTcpFinsHeader() *tcpFinsHeader {
	return &tcpFinsHeader{
		Magic:     [4]byte{'F', 'I', 'N', 'S'},
		Command:   TcpCommandFrameSend,
		ErrorCode: 0,
	}
}

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

	header := newTcpFinsHeader()
	header.Length = uint32(len(data)) + 8

	buf := &bytes.Buffer{}

	err = binary.Write(buf, binary.BigEndian, header)
	if err != nil {
		return 0, err
	}

	_, err = buf.Write(data)
	if err != nil {
		return 0, err
	}

	err = t.conn.SetWriteDeadline(time.Now().Add(t.WriteTimeout))
	if err != nil {
		return 0, err
	}

	return t.conn.Write(buf.Bytes())
}

func (t *TcpTransporter) ReadHeader() (header *respFinsHeader, err error) {
	defer func() {
		if err != nil {
			t.setState(StateDisconnected)
		}
	}()

	tcpHeaderBuf := make([]byte, 4*4)
	_, err = t.ReadData(tcpHeaderBuf)
	if err != nil {
		return nil, err
	}

	if !bytes.Equal(tcpHeaderBuf[:4], []byte("FINS")) {
		return nil, errors.New("invalid FINS header")
	}

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

func (t *TcpTransporter) ReadData(buf []byte) (n int, err error) {
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
