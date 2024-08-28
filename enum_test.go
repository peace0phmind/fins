package fins

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestEncodeAddress(t *testing.T) {
	addr, err := PlcTypeNew.EncodeAddress(&FinAddress{AreaCIO.MustType(DataTypeBit), 10, 13})

	assert.NoError(t, err, "EncodeAddress")
	assert.Equal(t, [4]byte{0x30, 0x0, 0x0a, 0x0d}, addr)
}
