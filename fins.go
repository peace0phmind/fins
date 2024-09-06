package fins

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"github.com/expgo/factory"
	"github.com/expgo/log"
	"github.com/expgo/structure"
	"sync/atomic"
)

/*
DataClass

	@Enum {
		Command
		Response
	}
*/
type DataClass int

type finsHeader struct {
	ICF byte
	RSV byte
	GCT byte
	DNA byte
	DA1 byte
	DA2 byte
	SNA byte
	SA1 byte
	SA2 byte
	SID byte
}

const respHeaderSize = 14

type respFinsHeader struct {
	finsHeader
	CommandCode [2]byte
	EndCode     EndCode
}

func newFinsHeader(dc DataClass, requireResp bool, sid byte) *finsHeader {
	ret := &finsHeader{}

	// bit7: bridge always 1
	ret.ICF = ret.ICF | 0b10000000
	// bit6: Data classification (0: Command; 1: Response)
	if dc == DataClassResponse {
		ret.ICF = ret.ICF | 0b01000000
	}
	// bit5-bit1, skip
	// bit0: Response (0: Required; 1: Not required)
	if !requireResp {
		ret.ICF = ret.ICF | 0b00000001
	}

	// Set GCT: fix 2 or 7 if across up to 8 network layers
	ret.GCT = 2

	/*
		Destination network address. Specify within the following ranges (hex).
		00: 		Local network
		01 to 7F: 	Remote network address (decimal: 1 to 127)
	*/
	ret.DNA = 0x00 // set to Local network

	/*
		Destination node address. Specify within the following ranges (hex).
		00: 			Internal communications in local PLC
		01 to 20:		Node address in Controller Link Network (1 to 32 decimal)
		01 to FE: FF: 	Ethernet (1 to 254 decimal, for Ethernet Units with model numbers ending in ETN21)
		DA2:			Broadcast transmission
	*/
	ret.DA1 = 0x00

	/*
		Destination unit address. Specify within the following ranges (hex).
		00: 		CPU Unit
		FE: 		Controller Link Unit or Ethernet Unit connected to network
		10 to 1F:	CPU Bus Unit
		E1:			Inner Board
	*/
	ret.DA2 = 0x00

	/*
		Source network address. Specify within the following ranges (hex).
		00: 		Local network
		01 to 7F: 	Remote network (1 to 127 decimal)
	*/
	ret.SNA = 0x00

	/*
		Source node address. Specify within the following ranges (hex).
		00: 		Internal communications in PLC
		01 to 20: 	Node address in Controller Link Network (1 to 32 decimal)
		01 to FE: 	Ethernet (1 to 254 decimal, for Ethernet Units with model numbers ending in ETN21)
	*/
	ret.SA1 = 0x00

	/*
		Source unit address. Specify within the following ranges (hex).
		00: 		CPU Unit
		10 to 1F: 	CPU Bus Unit
	*/
	ret.SA2 = 0x00 // set to CPU Unit

	/*
		Service ID. Used to identify the process generating the transmission.
		Set the SID to any number between 00 and FF
	*/
	ret.SID = sid

	return ret
}

type fins struct {
	log.InnerLog
	plcType     PlcType
	transporter Transporter
	sid         atomic.Uint32
}

func NewFins(plcType PlcType, transType TransType, addr string) Fins {
	ret := factory.New[fins]()

	ret.plcType = plcType

	switch transType {
	case TransTypeTcp:
		ret.transporter = newTcpTransport(addr)
	case TransTypeUdp:
		ret.transporter = newUdpTransport(addr)
	default:
		panic("unknown transporter type")
	}

	return ret
}

func (f *fins) Open() error {
	if f.transporter != nil {
		return f.transporter.Open()
	}

	return nil
}

func (f *fins) Close() error {
	if f.transporter != nil {
		defer func() {
			f.transporter = nil
		}()

		return f.transporter.Close()
	}

	return nil
}

func (f *fins) SetStateChangeCallback(callback func(oldState, newState State)) {
	if f.transporter != nil {
		f.transporter.SetStateChangeCallback(callback)
	}
}

func (f *fins) Read(address *FinAddress, length uint16) ([]*FinValue, error) {
	if length == 0 {
		return nil, errors.New("fins: Read called with zero length")
	}

	reqHeader := newFinsHeader(DataClassCommand, true, byte(f.sid.Add(1)))

	req := &bytes.Buffer{}
	_ = req.WriteByte(CommandMemoryRead.Mr())
	_ = req.WriteByte(CommandMemoryRead.Sr())

	addr, err := f.plcType.EncodeAddress(address)
	if err != nil {
		f.L.Warnf("failed to encode address: %v", err)
		return nil, err
	}

	_, _ = req.Write(addr[:])
	_ = binary.Write(req, binary.BigEndian, length)

	_, err = f.transporter.Write(reqHeader, req.Bytes())
	if err != nil {
		f.L.Warnf("write to transporter failed: %v", err)
		return nil, err
	}

	// read resp
	respHeader, err := f.transporter.ReadHeader()
	if err != nil {
		f.L.Warnf("read header from transporter failed: %v", err)
		return nil, err
	}

	if reqHeader.SID != respHeader.SID {
		f.transporter.setState(StateDisconnected)
		f.L.Error("req sid not equal to resp sid, reconnect to remote")
		return nil, fmt.Errorf("expected sid %v but got %v", respHeader.SID, reqHeader.SID)
	}

	if respHeader.CommandCode[0] != CommandMemoryRead.Mr() || respHeader.CommandCode[1] != CommandMemoryRead.Sr() {
		return nil, fmt.Errorf("invalid command: %x: %x", respHeader.CommandCode[0], respHeader.CommandCode[1])
	}

	err = respHeader.EndCode.Error()
	if err != nil {
		f.L.Warnf("end code failed: %v", err)
		return nil, err
	}

	itemSize := address.AreaCode.Size()
	resp := make([]byte, itemSize*int(length))
	_, err = f.transporter.ReadData(resp)
	if err != nil {
		f.L.Warnf("read data from transporter failed: %v", err)
		return nil, err
	}

	values := make([]*FinValue, length)
	for i := 0; i < int(length); i++ {
		newAddr := structure.Clone(address)
		newAddr.Address += uint16(i)
		values[i] = &FinValue{
			FinAddress: newAddr,
			Buf:        resp[i*itemSize : (i+1)*itemSize],
		}
	}

	return values, nil
}

func (f *fins) Write(address *FinAddress, values []*FinValue) error {
	if len(values) == 0 {
		return errors.New("no values to write")
	}

	reqHeader := newFinsHeader(DataClassCommand, true, byte(f.sid.Add(1)))

	req := &bytes.Buffer{}
	_ = req.WriteByte(CommandMemoryWrite.Mr())
	_ = req.WriteByte(CommandMemoryWrite.Sr())

	addr, err := f.plcType.EncodeAddress(address)
	if err != nil {
		f.L.Warnf("failed to encode address: %v", err)
		return err
	}

	_, _ = req.Write(addr[:])
	_ = binary.Write(req, binary.BigEndian, uint16(len(values)))

	for _, value := range values {
		req.Write(value.Buf)
	}

	_, err = f.transporter.Write(reqHeader, req.Bytes())
	if err != nil {
		f.L.Warnf("write to transporter failed: %v", err)
		return err
	}

	// read resp
	respHeader, err := f.transporter.ReadHeader()
	if err != nil {
		f.L.Warnf("read from transporter failed: %v", err)
		return err
	}

	if reqHeader.SID != respHeader.SID {
		f.transporter.setState(StateDisconnected)
		f.L.Error("req sid not equal to resp sid, reconnect to remote")
		return fmt.Errorf("expected sid %v but got %v", respHeader.SID, reqHeader.SID)
	}

	if respHeader.CommandCode[0] != CommandMemoryWrite.Mr() || respHeader.CommandCode[1] != CommandMemoryWrite.Sr() {
		return fmt.Errorf("invalid command: %x: %x", respHeader.CommandCode[0], respHeader.CommandCode[1])
	}

	err = respHeader.EndCode.Error()
	if err != nil {
		f.L.Warnf("end code failed: %v", err)
	}

	return err
}

func (f *fins) RandomRead(addresses []*FinAddress) ([]*FinValue, error) {
	if len(addresses) == 0 {
		return nil, errors.New("no addresses to read")
	}

	reqHeader := newFinsHeader(DataClassCommand, true, byte(f.sid.Add(1)))

	req := &bytes.Buffer{}
	_ = req.WriteByte(CommandMultipleMemoryRead.Mr())
	_ = req.WriteByte(CommandMultipleMemoryRead.Sr())

	itemsSize := 0

	for _, address := range addresses {
		addr, err := f.plcType.EncodeAddress(address)
		if err != nil {
			f.L.Warnf("failed to encode address: %v", err)
			return nil, err
		}

		itemsSize += address.AreaCode.Size()

		req.Write(addr[:])
	}

	_, err := f.transporter.Write(reqHeader, req.Bytes())
	if err != nil {
		f.L.Warnf("write to transporter failed: %v", err)
		return nil, err
	}

	// read resp
	respHeader, err := f.transporter.ReadHeader()
	if err != nil {
		f.L.Warnf("read from transporter failed: %v", err)
		return nil, err
	}

	if reqHeader.SID != respHeader.SID {
		f.transporter.setState(StateDisconnected)
		f.L.Error("req sid not equal to resp sid, reconnect to remote")
		return nil, fmt.Errorf("expected sid %v but got %v", respHeader.SID, reqHeader.SID)
	}

	if respHeader.CommandCode[0] != CommandMultipleMemoryRead.Mr() || respHeader.CommandCode[1] != CommandMultipleMemoryRead.Sr() {
		return nil, fmt.Errorf("invalid command: %x: %x", respHeader.CommandCode[0], respHeader.CommandCode[1])
	}

	err = respHeader.EndCode.Error()
	if err != nil {
		f.L.Warnf("end code failed: %v", err)
		return nil, err
	}

	resp := make([]byte, itemsSize+len(addresses))
	_, err = f.transporter.ReadData(resp)
	if err != nil {
		f.L.Warnf("read data from transporter failed: %v", err)
		return nil, err
	}

	var values []*FinValue
	readSize := 0
	for _, address := range addresses {
		itemSize := address.AreaCode.Size()
		value := &FinValue{
			FinAddress: address,
			Buf:        resp[readSize+1 : readSize+1+itemSize],
		}
		readSize += 1 + itemSize
		values = append(values, value)
	}

	return values, nil
}
