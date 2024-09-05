package fins

import (
	"encoding/binary"
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

func (fv *FinValue) Uint16() uint16 {
	return fv.Value().(uint16)
}

func (fv *FinValue) Uint32() uint32 {
	return fv.Value().(uint32)
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
	Open() error
	Close() error
	Read(address *FinAddress, length uint16) ([]*FinValue, error)
	Write(address *FinAddress, values []*FinValue) error
	RandomRead(addresses []*FinAddress) ([]*FinValue, error)
}
