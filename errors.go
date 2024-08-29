package fins

import "errors"

type EndCode [2]byte

func (e EndCode) NetWorkRelayError() bool {
	return e[0]&(1<<7) != 0
}

func (e EndCode) FatalCpuUnitError() bool {
	return e[1]&(1<<7) != 0
}

func (e EndCode) NonFatalCpuUnitError() bool {
	return e[1]&(1<<6) != 0
}

func (e EndCode) MainCode() byte {
	return e[0] & 0b1111111
}

func (e EndCode) SubCode() byte {
	return e[1] & 0b111111
}

var (
	NetWorkRelayError    = errors.New("network relay error")
	FatalCpuUnitError    = errors.New("fatal cpu unit error")
	NonFatalCpuUnitError = errors.New("non-fatal cpu unit error")
)

/*
MC

	@Enum {
		NormalCompletion 			= 0x00
		LocalNodeError 				= 0x01
		DestinationNodeError		= 0x02
		ControllerError				= 0x03
		ServiceUnsupported			= 0x04
		RoutingTableError			= 0x05
		CommandFormatError			= 0x10
		ParameterError				= 0x11
		ReadNotPossible				= 0x20
		WriteNotPossible			= 0x21
		NotExecutableInCurrentMode	= 0x22
		NoSuchDevice				= 0x23
		CannotStartStop				= 0x24
		UnitError					= 0x25
		CommandError				= 0x26
		AccessRightError			= 0x30
		Abort						= 0x40
	}
*/
type MC byte

var errorsMap = map[MC]map[byte]string{
	MCNormalCompletion: {
		0x01: "Service canceled",
	},
	MCLocalNodeError: {
		0x01: "Local node not in network",
		0x02: "Token timeout",
		0x03: "Retries failed",
		0x04: "Too many send frames",
		0x05: "Node address range error",
		0x06: "Node address duplication",
	},
	MCDestinationNodeError: {
		0x01: "Destination node not in network",
		0x02: "Unit missing",
		0x03: "Third node missing",
		0x04: "Destination node busy",
		0x05: "Response timeout",
	},
	MCControllerError: {
		0x01: "Communications controller error",
		0x02: "CPU Unit error",
		0x03: "Controller error",
		0x04: "Unit number error",
	},
	MCServiceUnsupported: {
		0x01: "Undefined command",
		0x02: "Not supported by model/version",
	},
	MCRoutingTableError: {
		0x01: "Destination address setting error",
		0x02: "No routing tables",
		0x03: "Routing table error",
		0x04: "Too many relays",
	},
	MCCommandFormatError: {
		0x01: "Command too long",
		0x02: "Command too short",
		0x03: "Elements/data don’t match",
		0x04: "Command format error",
		0x05: "Header error",
	},
	MCParameterError: {
		0x01: "Area classification missing",
		0x02: "Access size error",
		0x03: "Address range error",
		0x04: "Address range exceeded",
		0x06: "Program missing",
		0x09: "Relational error",
		0x0A: "Duplicate data access",
		0x0B: "Response too long",
		0x0C: "Parameter error",
	},
	MCReadNotPossible: {
		0x02: "Protected",
		0x03: "Table missing",
		0x04: "Data missing",
		0x05: "Program missing",
		0x06: "File missing",
		0x07: "Data mismatch",
	},
	MCWriteNotPossible: {
		0x01: "Read-only",
		0x02: "Protected: Cannot write data link table",
		0x03: "Cannot register",
		0x05: "Program missing",
		0x06: "File missing",
		0x07: "File name already exists",
		0x08: "Cannot change",
	},
	MCNotExecutableInCurrentMode: {
		0x01: "Not possible during execution",
		0x02: "Not possible while running",
		0x03: "Wrong PLC mode, The PLC is in PROGRAM mode.",
		0x04: "Wrong PLC mode, The PLC is in DEBUG mode.",
		0x05: "Wrong PLC mode, The PLC is in MONITOR mode.",
		0x06: "Wrong PLC mode, The PLC is in RUN mode.",
		0x07: "Specified node not polling node",
		0x08: "Step cannot be executed",
	},
	MCNoSuchDevice: {
		0x01: "File device missing",
		0x02: "Memory missing",
		0x03: "Clock missing",
	},
	MCCannotStartStop: {
		0x01: "Table missing",
	},
	MCUnitError: {
		0x02: "Memory error",
		0x03: "I/O setting error",
		0x04: "Too many I/O points",
		0x05: "CPU bus error",
		0x06: "I/O duplication",
		0x07: "I/O bus error",
		0x09: "SYSMAC BUS/2 error",
		0x0A: "CPU Bus Unit error",
		0x0D: "SYSMAC BUS No. duplication",
		0x0F: "Memory error",
		0x10: "SYSMAC BUS terminator missing",
	},
	MCCommandError: {
		0x01: "No protection",
		0x02: "Incorrect password",
		0x04: "Protected",
		0x05: "Service already executing",
		0x06: "Service stopped",
		0x07: "No execution right",
		0x08: "Settings not complete",
		0x09: "Necessary items not set",
		0x0A: "Number already defined",
		0x0B: "Error will not clear",
	},
	MCAccessRightError: {
		0x01: "No access right",
	},
	MCAbort: {
		0x01: "Service aborted",
	},
}

func (e EndCode) Error() error {
	if e[0] == 0 && e[1] == 0 {
		return nil
	}

	if e.NetWorkRelayError() {
		return NetWorkRelayError
	}

	if e.FatalCpuUnitError() {
		return FatalCpuUnitError
	}

	if e.NonFatalCpuUnitError() {
		return NonFatalCpuUnitError
	}

	return nil
}
