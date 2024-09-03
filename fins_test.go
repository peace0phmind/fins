package fins

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestFins(t *testing.T) {
	f := NewFins(PlcTypeNew, TransTypeTcp, "")

	err := f.Open()
	assert.NoError(t, err)

	//f.Read(&FinAddress{AreaCode: MemoryArea})
}
