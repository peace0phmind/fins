package fins

type FinAddress struct {
	AreaCode MemoryArea
	Address  uint16
	Offset   byte
}

type FinValue struct {
	*FinAddress
	Value any
}

type Fins interface {
	Read(address *FinAddress, length uint16) ([]*FinValue, error)
	Write(address *FinAddress, values []*FinValue) error
	RandomRead(addresses []*FinAddress) ([]*FinValue, error)
}

type fins struct {
}

func (f *fins) Read(address *FinAddress, length uint16) ([]*FinValue, error) {
	return nil, nil
}

func (f *fins) Write(address *FinAddress, values []*FinValue) error {
	return nil
}

func (f *fins) RandomRead(addresses []*FinAddress) ([]*FinValue, error) {
	return nil, nil
}
