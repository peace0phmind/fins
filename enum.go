package fins

/*
MemoryArea

	@EnumConfig(NoCamel)
	@Enum(size int, areaName string, dateType string) {
		CIOBit(1, "CIO", "Bit") = 0x30  	// CIO Area
		WBit(1, "W", "Bit") = 0x31			// Work Area
		HBit(1, "H", "Bit") = 0x32			// Holding Bit Area
		ABit(1, "A", "Bit") = 0x33			// Auxiliary Bit Area
		CIOWord(2, "CIO", "Word") = 0xB0	// CIO Area
		WWord(2, "W", "Word") = 0xB1		// Work Area
		HWord(2, "H", "Word") = 0xB2		// Holding Bit Area
		AWord(2, "A", "Word") = 0xB3		// Auxiliary Bit Area
		TBit(1, "T", "CF") = 0x09			// Timer Area
		//CBit(1, "C", "CF") = 0x09			// Counter Area
		TWord(2, "T", "PV") = 0x89			// Timer Area
		//CWord(2, "C", "PV") = 0x89			// Counter Area
		DBit(1, "D", "Bit") = 0x02			// DM Area
		DWord(2, "D", "Word") = 0x82		// DM Area
		IR(4, "IR", "PV") = 0xDC			// Index Register
		DR(2, "DR", "PV") = 0xBC			// Data Register
	}
*/
type MemoryArea byte

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
