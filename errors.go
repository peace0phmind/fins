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
