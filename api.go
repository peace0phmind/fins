package fins

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"github.com/expgo/factory"
	"github.com/expgo/log"
	"github.com/expgo/structure"
)

type FinAddress struct {
	AreaCode MemoryArea
	Address  uint16
	Offset   byte
}

type FinValue struct {
	*FinAddress
	Buf []byte
}

func (fv *FinValue) Value() any {
	switch fv.AreaCode.Size() {
	case 1:
		return fv.Buf[0]
	case 2:
		return binary.BigEndian.Uint16(fv.Buf)
	case 4:
		return binary.BigEndian.Uint32(fv.Buf)
	default:
		return nil
	}
}

func (fv *FinValue) SetValue(value any) error {
	switch fv.AreaCode.Size() {
	case 1:
		fv.Buf = []byte{structure.MustConvertTo[byte](value)}
	case 2:
		fv.Buf = binary.BigEndian.AppendUint16(nil, structure.MustConvertTo[uint16](value))
	case 4:
		fv.Buf = binary.BigEndian.AppendUint32(nil, structure.MustConvertTo[uint32](value))
	}

	return nil
}

type Fins interface {
	Read(address *FinAddress, length uint16) ([]*FinValue, error)
	Write(address *FinAddress, values []*FinValue) error
	RandomRead(addresses []*FinAddress) ([]*FinValue, error)
}

type fins struct {
	log.InnerLog
	plcType     PlcType
	transporter Transporter
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

func (f *fins) Read(address *FinAddress, length uint16) ([]*FinValue, error) {
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

	_, err = f.transporter.Write(req.Bytes())
	if err != nil {
		f.L.Warnf("write to transporter failed: %v", err)
		return nil, err
	}

	// read resp
	itemSize := address.AreaCode.Size()
	respSize := 4 + itemSize*int(length)

	resp := make([]byte, respSize)
	_, err = f.transporter.Read(resp)
	if err != nil {
		f.L.Warnf("read from transporter failed: %v", err)
		return nil, err
	}

	if resp[0] != CommandMemoryRead.Mr() || resp[1] != CommandMemoryRead.Sr() {
		return nil, fmt.Errorf("invalid command: %x: %x", resp[0], resp[1])
	}

	endCode := EndCode{resp[2], resp[3]}
	err = endCode.Error()
	if err != nil {
		f.L.Warnf("end code failed: %v", err)
		return nil, err
	}

	values := make([]*FinValue, length)
	for i := 0; i < int(length); i++ {
		newAddr := structure.Clone(address)
		newAddr.Address += 1
		values[i] = &FinValue{
			FinAddress: newAddr,
			Buf:        resp[i*itemSize+4 : (i+1)*itemSize+4],
		}
	}

	return values, nil
}

func (f *fins) Write(address *FinAddress, values []*FinValue) error {
	if len(values) == 0 {
		return errors.New("no values to write")
	}

	req := &bytes.Buffer{}

	_ = req.WriteByte(CommandMemoryWrite.Mr())
	_ = req.WriteByte(CommandMemoryWrite.Sr())

	addr, err := f.plcType.EncodeAddress(address)
	if err != nil {
		f.L.Warnf("failed to encode address: %v", err)
		return err
	}

	_, _ = req.Write(addr[:])
	_ = binary.Write(req, binary.BigEndian, len(values))

	for _, value := range values {
		req.Write(value.Buf)
	}

	_, err = f.transporter.Write(req.Bytes())
	if err != nil {
		f.L.Warnf("write to transporter failed: %v", err)
		return err
	}

	// read resp
	resp := make([]byte, 4)
	_, err = f.transporter.Read(resp)
	if err != nil {
		f.L.Warnf("read from transporter failed: %v", err)
		return err
	}

	if resp[0] != CommandMemoryWrite.Mr() || resp[1] != CommandMemoryWrite.Sr() {
		return fmt.Errorf("invalid command: %x: %x", resp[0], resp[1])
	}

	endCode := EndCode{resp[2], resp[3]}
	err = endCode.Error()
	if err != nil {
		f.L.Warnf("end code failed: %v", err)
	}

	return err
}

func (f *fins) RandomRead(addresses []*FinAddress) ([]*FinValue, error) {
	if len(addresses) == 0 {
		return nil, errors.New("no addresses to read")
	}

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

	_, err := f.transporter.Write(req.Bytes())
	if err != nil {
		f.L.Warnf("write to transporter failed: %v", err)
		return nil, err
	}

	// read resp
	resp := make([]byte, 4+itemsSize)
	_, err = f.transporter.Read(resp)
	if err != nil {
		f.L.Warnf("read from transporter failed: %v", err)
		return nil, err
	}

	if resp[0] != CommandMultipleMemoryRead.Mr() || resp[1] != CommandMultipleMemoryRead.Sr() {
		return nil, fmt.Errorf("invalid command: %x: %x", resp[0], resp[1])
	}

	endCode := EndCode{resp[2], resp[3]}
	err = endCode.Error()
	if err != nil {
		f.L.Warnf("end code failed: %v", err)
		return nil, err
	}

	values := make([]*FinValue, len(addresses))
	readSize := 4
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
