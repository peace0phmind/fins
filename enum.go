package fins

/*
PlcType

	@Enum (description string){
		New("CS/CJ/CP/NSJ-series")
		Old("CVM1/CV-series")
	}
*/
type PlcType int

/*
MemoryArea word is big end

	@EnumConfig(NoCamel)
	@Enum(areaName string, dateType string, code byte, max int, offset int, oldCode byte, oldMax int, oldOffset int, size int) {
		CIOBit("CIO", "Bit", 0x30, 6143, 0, 0x0, 2555, 0, 1)   			// CIO Area Bit
		WBit("WR", "Bit", 0x31, 511, 0, 0x0, -1, -1, 1) 		 			// Work Area Bit
		HBit("HR", "Bit", 0x32, 511, 0, 0x0, -1, -1, 1) 		 			// Holding Bit Area Bit
		ABit("AR", "Bit", 0x33, 959, 0, 0x0, 959, 0x0B00, 1) 	 		// Auxiliary Bit Area Bit

		CIOBitFs("CIO", "BitFs", 0x70, 6143, 0, 0x40, 2555, 0, 1)   	// CIO Area Bit with forced status
		WBitFs("WR", "BitFs", 0x71, 511, 0, 0x0, -1, -1, 1) 		 		// Work Area Bit with forced status
		HBitFs("HR", "BitFs", 0x72, 511, 0, 0x0, -1, -1, 1) 		 		// Holding Bit Area Bit with forced status

		CIOWord("CIO", "Word", 0xB0, 6143, 0, 0x80, 2555, 0, 2)  		// CIO Area Word
		WWord("WR", "Word", 0xB1, 511, 0, 0x0, -1, -1, 2) 	 			// Work Area Word
		HWord("HR", "Word", 0xB2, 511, 0, 0x0, -1, -1, 2) 	 			// Holding Bit Area Word
		AWord("AR", "Word", 0xB3, 959, 0, 0x80, 959, 0x0B00, 2) 	 	// Auxiliary Bit Area Word

		CIOWordFs("CIO", "WordFs", 0xF0, 6143, 0, 0xC0, 2555, 0, 2) 	// CIO Area Word with forced status
		WWordFs("WR", "WordFs", 0xF1, 511, 0, 0x, -1, -1, 2) 	 			// Work Area Word with forced status
		HWordFs("HR", "WordFs", 0xF2, 511, 0, 0x, -1, -1, 2) 	 			// Holding Bit Area Word with forced status

		TBit("TIM", "CF", 0x09, 4095, 0, 0x01, 2047, 0, 1) 	 			// Timer Area
		CBit("CNT", "CF", 0x09, 4095, 0x8000, 0x01, 2047, 0x0800, 1)	// Counter Area

		TBitFs("TIM", "CFFs", 0x49, 4095, 0, 0x41, 2047, 0, 1) 	 		// Timer Area with forced status
		CBitFs("CNT", "CFFs", 0x49, 4095, 0x8000, 0x41, 2047, 0x0800, 1)// Counter Area with forced status

		TWord("TIM", "PV", 0x89, 4095, 0, 0x81, 2047, 0, 2) 	 		// Timer Area
		CWord("CNT", "PV", 0x89, 4095, 0x8000, 0x81, 2047, 0x0800, 2) 	// Counter Area

		DBit("DM", "Bit", 0x02, 32767, 0, 0x0, -1, -1, 1) 	 			// DM Area
		DWord("DM", "Word", 0x82, 32767, 0, 0x82, 32767, 0, 2)  		// DM Area

		IR("IR", "PV", 0xDC, 15, 0x0100, 0x0, -1, -1, 4)  				// Index Register
		DR("DR", "PV", 0xBC, 15, 0x0200, 0x9C, 2, 0x03, 2) 		 			// Data Register
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
