package fins

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestFins(t *testing.T) {
	f := NewFins(PlcTypeNew, TransTypeTcp, "0.0.0.0:9600")

	err := f.Open()
	assert.NoError(t, err)

	ret, err := f.Read(&FinAddress{AreaCode: MemoryAreaWRWord, Address: 0, Offset: 0}, 0)
	assert.NoError(t, err)

	println(ret[0].Value())
}
