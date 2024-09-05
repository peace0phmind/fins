package fins

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestFinsRead(t *testing.T) {
	f := NewFins(PlcTypeNew, TransTypeTcp, "0.0.0.0:9600")

	err := f.Open()
	defer func() {
		_ = f.Close()
	}()

	assert.NoError(t, err)

	ret, err := f.Read(&FinAddress{AreaCode: MemoryAreaDMWord, Address: 0, Offset: 0}, 1)
	assert.NoError(t, err)

	if err == nil {
		println(ret[0].Uint16())
	}
}

func TestFinsWrite(t *testing.T) {
	f := NewFins(PlcTypeNew, TransTypeTcp, "0.0.0.0:9600")

	err := f.Open()
	defer func() {
		_ = f.Close()
	}()

	assert.NoError(t, err)

	addr := &FinAddress{AreaCode: MemoryAreaDMWord, Address: 0, Offset: 0}
	value := &FinValue{FinAddress: addr}
	_ = value.SetValue(uint16(8))
	err = f.Write(addr, []*FinValue{value})
	assert.NoError(t, err)
}

func TestFinsRandomRead(t *testing.T) {
	f := NewFins(PlcTypeNew, TransTypeTcp, "0.0.0.0:9600")

	err := f.Open()
	defer func() {
		_ = f.Close()
	}()

	assert.NoError(t, err)

	values, err := f.RandomRead([]*FinAddress{{AreaCode: MemoryAreaDMWord, Address: 0}, {AreaCode: MemoryAreaWRWord, Address: 0}})
	assert.NoError(t, err)

	if err == nil {
		println(values[0].Uint16())
		println(values[1].Uint16())
	}
}

func TestFinsReadUdp(t *testing.T) {
	f := NewFins(PlcTypeNew, TransTypeUdp, "192.168.1.232:59600")

	err := f.Open()
	defer func() {
		_ = f.Close()
	}()

	assert.NoError(t, err)

	ret, err := f.Read(&FinAddress{AreaCode: MemoryAreaDMWord, Address: 0, Offset: 0}, 1)
	assert.NoError(t, err)

	if err == nil {
		println(ret[0].Uint16())
	}
}

func TestFinsWriteUdp(t *testing.T) {
	f := NewFins(PlcTypeNew, TransTypeUdp, "192.168.1.232:59600")

	err := f.Open()
	defer func() {
		_ = f.Close()
	}()

	assert.NoError(t, err)

	addr := &FinAddress{AreaCode: MemoryAreaDMWord, Address: 0, Offset: 0}
	value := &FinValue{FinAddress: addr}
	_ = value.SetValue(uint16(8))
	err = f.Write(addr, []*FinValue{value})
	assert.NoError(t, err)
}

func TestFinsRandomReadUdp(t *testing.T) {
	f := NewFins(PlcTypeNew, TransTypeUdp, "192.168.1.232:59600")

	err := f.Open()
	defer func() {
		_ = f.Close()
	}()

	assert.NoError(t, err)

	values, err := f.RandomRead([]*FinAddress{{AreaCode: MemoryAreaDMWord, Address: 0}, {AreaCode: MemoryAreaWRWord, Address: 0}})
	assert.NoError(t, err)

	if err == nil {
		println(values[0].Uint16())
		println(values[1].Uint16())
	}
}
