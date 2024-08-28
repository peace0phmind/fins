package fins

import "errors"

/*
PlcType

	@Enum (description string){
		New("CS/CJ/CP/NSJ-series")
		Old("CVM1/CV-series")
	}
*/
type PlcType int

/*
DataType

	@EnumConfig(NoCamel)
	@Enum {
		Bit
		BitFs
		Word
		WordFs
		CF
		CFFs
		PV
	}
*/
type DataType string

/*
Area

	@EnumConfig(NoCamel, NoCase)
	@Enum {
		CIO
		WR
		HR
		AR
		TIM
		CNT
		DM
		IR
		DR
	}
*/
type Area string

func (a Area) WithType(dataType DataType) (MemoryArea, error) {
	return ParseMemoryArea(a.Val() + dataType.Val())
}

func (a Area) MustType(dataType DataType) MemoryArea {
	return MustParseMemoryArea(a.Val() + dataType.Val())
}

/*
MemoryArea word is big end

	@EnumConfig(NoCamel, NoCase, MustParse)
	@Enum(areaName string, dataType string, code byte, max uint16, offset uint16, oldCode byte, oldMax uint16, oldOffset uint16, size int) {
		CIOBit("CIO", "Bit", 0x30, 6143, 0, 0x0, 2555, 0, 1)   			// CIO Area Bit
		WRBit("WR", "Bit", 0x31, 511, 0, 0x0, -1, -1, 1) 		 		// Work Area Bit
		HRBit("HR", "Bit", 0x32, 511, 0, 0x0, -1, -1, 1) 		 		// Holding Bit Area Bit
		ARBit("AR", "Bit", 0x33, 959, 0, 0x0, 959, 0x0B00, 1) 	 		// Auxiliary Bit Area Bit

		CIOBitFs("CIO", "BitFs", 0x70, 6143, 0, 0x40, 2555, 0, 1)   	// CIO Area Bit with forced status
		WRBitFs("WR", "BitFs", 0x71, 511, 0, 0x0, -1, -1, 1) 		 	// Work Area Bit with forced status
		HRBitFs("HR", "BitFs", 0x72, 511, 0, 0x0, -1, -1, 1) 		 	// Holding Bit Area Bit with forced status

		CIOWord("CIO", "Word", 0xB0, 6143, 0, 0x80, 2555, 0, 2)  		// CIO Area Word
		WRWord("WR", "Word", 0xB1, 511, 0, 0x0, -1, -1, 2) 	 			// Work Area Word
		HRWord("HR", "Word", 0xB2, 511, 0, 0x0, -1, -1, 2) 	 			// Holding Bit Area Word
		ARWord("AR", "Word", 0xB3, 959, 0, 0x80, 959, 0x0B00, 2) 	 	// Auxiliary Bit Area Word

		CIOWordFs("CIO", "WordFs", 0xF0, 6143, 0, 0xC0, 2555, 0, 4) 	// CIO Area Word with forced status
		WRWordFs("WR", "WordFs", 0xF1, 511, 0, 0x0, -1, -1, 4) 	 		// Work Area Word with forced status
		HRWordFs("HR", "WordFs", 0xF2, 511, 0, 0x0, -1, -1, 4) 	 		// Holding Bit Area Word with forced status

		TIMCF("TIM", "CF", 0x09, 4095, 0, 0x01, 2047, 0, 1) 	 			// Timer Area
		CNTCF("CNT", "CF", 0x09, 4095, 0x8000, 0x01, 2047, 0x0800, 1)	// Counter Area

		TIMCFFs("TIM", "CFFs", 0x49, 4095, 0, 0x41, 2047, 0, 1) 	 		// Timer Area with forced status
		CNTCFFs("CNT", "CFFs", 0x49, 4095, 0x8000, 0x41, 2047, 0x0800, 1)// Counter Area with forced status

		TIMPV("TIM", "PV", 0x89, 4095, 0, 0x81, 2047, 0, 2) 	 		// Timer Area
		CNTPV("CNT", "PV", 0x89, 4095, 0x8000, 0x81, 2047, 0x0800, 2) 	// Counter Area

		DMBit("DM", "Bit", 0x02, 32767, 0, 0x0, -1, -1, 1) 	 			// DM Area
		DMWord("DM", "Word", 0x82, 32767, 0, 0x82, 32767, 0, 2)  		// DM Area

		IRPV("IR", "PV", 0xDC, 15, 0x0100, 0x0, -1, -1, 4)  				// Index Register
		DRPV("DR", "PV", 0xBC, 15, 0x0200, 0x9C, 2, 0x03, 2) 		 		// Data Register
	}
*/
type MemoryArea string

/*
Command

	@Enum(mr byte, sr byte) {
		MemoryRead(1, 1)
		MemoryWrite(1, 2)
		MemoryFill(1, 3)
		MultipleMemoryRead(1, 4)
		MemoryTransfer(1, 5)
	}
*/
type Command int

func (pt PlcType) EncodeAddress(address *FinAddress) (ret [4]byte, err error) {
	ac := address.AreaCode
	if ac.DataType() == DataTypeBit.Val() {
		if address.Offset > 15 {
			return ret, errors.New("offset out of range")
		}
	} else {
		if address.Offset > 0 {
			return ret, errors.New("offset must be 0")
		}
	}

	switch pt {
	case PlcTypeNew:
		if address.Address > ac.Max() {
			return ret, errors.New("address out of range")
		}

		ret[0] = ac.Code()
		addr := address.Address + ac.Offset()
		ret[1] = byte(addr >> 8)
		ret[2] = byte(addr)
		ret[3] = address.Offset

		return

	case PlcTypeOld:
		if address.Address > ac.OldMax() {
			return ret, errors.New("address out of range")
		}

		ret[0] = ac.OldCode()
		addr := address.Address + ac.OldOffset()
		ret[1] = byte(addr >> 8)
		ret[2] = byte(addr)
		ret[3] = address.Offset

		return

	default:
		return ret, errors.New("invalid PlcType")
	}
}
