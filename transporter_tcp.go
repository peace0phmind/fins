package fins

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"github.com/expgo/factory"
	"net"
	"time"
)

/*
TcpCommand

	@Enum {
		NodeAddressClientToServer 	= 0x0
		NodeAddressServerToClient 	= 0x1
		FrameSend 					= 0x2
	}
*/
type TcpCommand uint32

type tcpFinsHeader struct {
	Magic     [4]byte
	Length    uint32
	Command   TcpCommand
	ErrorCode uint32
}

func newTcpFinsHeader(cmd TcpCommand) *tcpFinsHeader {
	return &tcpFinsHeader{
		Magic:     [4]byte{'F', 'I', 'N', 'S'},
		Command:   cmd,
		ErrorCode: 0,
	}
}

type TcpTransporter struct {
	baseTransporter
	da1 byte
	sa1 byte
}

func newTcpTransport(addr string) *TcpTransporter {
	return factory.NewWithFunc[TcpTransporter](func() *TcpTransporter {
		return &TcpTransporter{baseTransporter: baseTransporter{addr: addr}}
	})
}

func (t *TcpTransporter) Open() (err error) {
	if !t.running.CompareAndSwap(false, true) {
		return nil
	}

	if t.state == StateConnected {
		return nil
	}

	t.setState(StateConnecting, nil)
	dialer := net.Dialer{Timeout: 3 * time.Second}
	t.conn, err = dialer.Dial("tcp", t.addr)
	if err != nil {
		t.L.Warnf("DialTCP %s failed: %v", t.addr, err)
		t.setState(StateDisconnected, err)
		return err
	}

	err = t.getDaSa()

	if err == nil {
		t.setState(StateConnected, nil)
	}

	return err
}

func (t *TcpTransporter) getDaSa() (err error) {
	defer func() {
		if err != nil {
			t.setState(StateDisconnected, err)
		}
	}()

	tcpHeader := newTcpFinsHeader(TcpCommandNodeAddressClientToServer)
	tcpHeader.Length = 12

	req := &bytes.Buffer{}
	err = binary.Write(req, binary.BigEndian, tcpHeader)
	if err != nil {
		return err
	}

	err = binary.Write(req, binary.BigEndian, int32(0))
	if err != nil {
		return err
	}

	_, err = t.conn.Write(req.Bytes())
	if err != nil {
		return err
	}

	respTcpHeader, err := t.ReadTcpHeader()
	if err != nil {
		return err
	}

	if respTcpHeader.Length != 16 {
		return errors.New("invalid tcp header length for Node Address Server to Client")
	}

	buf := make([]byte, 8)
	_, err = t.conn.Read(buf)
	if err != nil {
		return err
	}

	cna := binary.BigEndian.Uint32(buf[0:4])
	sna := binary.BigEndian.Uint32(buf[4:8])

	t.da1 = byte(sna)
	t.sa1 = byte(cna)

	return nil
}

func (t *TcpTransporter) Close() (err error) {
	defer func() {
		t.setState(StateConnectClosed, err)
		t.conn = nil
	}()

	_ = t.baseTransporter.Close()

	if t.conn == nil {
		return nil
	}

	return t.conn.Close()
}

func (t *TcpTransporter) Write(header *finsHeader, data []byte) (n int, err error) {
	if t.conn == nil || t.state == StateDisconnected {
		return 0, errors.New("tcp transporter not connected")
	}

	defer func() {
		if err != nil {
			t.setState(StateDisconnected, err)
		}
	}()

	header.DA1 = t.da1
	header.SA1 = t.sa1

	tcpHeader := newTcpFinsHeader(TcpCommandFrameSend)
	tcpHeader.Length = uint32(len(data)) + 18

	buf := &bytes.Buffer{}

	err = binary.Write(buf, binary.BigEndian, tcpHeader)
	if err != nil {
		return 0, err
	}

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

func (t *TcpTransporter) ReadTcpHeader() (tcpHeader *tcpFinsHeader, err error) {
	tcpHeaderBuf := make([]byte, 4*4)
	_, err = t.ReadData(tcpHeaderBuf)
	if err != nil {
		return nil, err
	}

	tcpHeader = &tcpFinsHeader{}
	if err = binary.Read(bytes.NewBuffer(tcpHeaderBuf), binary.BigEndian, tcpHeader); err != nil {
		return nil, err
	}

	if !bytes.Equal(tcpHeader.Magic[:], []byte("FINS")) {
		return nil, errors.New("invalid FINS header")
	}

	if tcpHeader.ErrorCode != 0 {
		if tcpHeader.Length > 8 {
			buf := make([]byte, tcpHeader.Length-8)
			_, _ = t.conn.Read(buf)
		}
		return nil, fmt.Errorf("FINS error code: %d", tcpHeader.ErrorCode)
	}

	return tcpHeader, nil
}

func (t *TcpTransporter) ReadHeader() (header *respFinsHeader, err error) {
	if t.conn == nil || t.state == StateDisconnected {
		return nil, errors.New("tcp transporter not connected")
	}

	defer func() {
		if err != nil {
			t.setState(StateDisconnected, err)
		}
	}()

	_, err = t.ReadTcpHeader()
	if err != nil {
		return nil, err
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
	if t.conn == nil || t.state == StateDisconnected {
		return 0, errors.New("tcp transporter not connected")
	}

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
