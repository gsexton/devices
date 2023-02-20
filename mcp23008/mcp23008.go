// An implementation for interacting with a Microchip MCP23008/MCP23S08
// 8 Bit IO Expander with I2C/SPI Interface. The chip provides 8 bits
// of GPIO via an I2C or SPI interface.
//
// The best source of information ont his chip is to refer to the datasheet.
//
package mcp23008

import (
    "log"
    "periph.io/x/conn"
)

type ChipRegister byte
type PinState byte
type OutputMode byte
type MCP23008 struct {
	conn conn.Conn
}

const (
	// IO Direction for pins. 0 = Output, 1 = Input
	IODIR ChipRegister = iota
	// Input Polarity Register - Bitmap, 0 = Normal, 1 = Inverted
	IPOL
	// Interrupt on Change Bit map. 0 = Disable interrupt, 1 = enable
	GPINTEN
	// Default Compare Register for interrupt on change.
	DEFVAL
	// Interrupt On Change Control - See Datasheet
	INTCON
	// IO Controls for Device Operation. See Datasheet for bitmap values.
	IOCON
	// Pullup Resistor Bitmask. 0=Pullup Disabled. 1 = Pullup Enabled.
	GPPU
	// Interrupt Register flag. Bit set for pins generating interrupt.
	INTF
	// Interrupt Capture  - Captures the value of the GPIO port when the
	// interrupt was triggered.
	INTCAP
	// The GPIO Port Values
	GPIO
	// Output Latch.
	OLAT
)

const (
	PinOutput PinState = iota
	PinInput
)

const (
	ModeOutput OutputMode = iota
	ModeInput
)

// A bitmap for turning on bits.
var maskOn = [...]byte{
	0b00000001,
	0b00000010,
	0b00000100,
	0b00001000,
	0b00010000,
	0b00100000,
	0b01000000,
	0b10000000,
}

// A bitmap for turn off bits.
var maskOff = [...]byte{
	0b11111110,
	0b11111101,
	0b11111011,
	0b11110111,
	0b11101111,
	0b11011111,
	0b10111111,
	0b01111111,
}

func NewMCP23008(conn conn.Conn) (*MCP23008, error) {
	mcp := MCP23008{conn: conn}
	initRegisters := [...]ChipRegister{IODIR, IPOL, GPINTEN, DEFVAL, GPPU, INTF, GPIO}
	var err error
	for _, register := range initRegisters {
		err := mcp.WriteRegister(register, 0)
        if err!=nil {
            return nil, err
        }
	}
	return &mcp, err
}

func (mcp *MCP23008) ReadGPIO() (byte, error) {
	return mcp.ReadRegister(GPIO)
}

func (mcp *MCP23008) String() string {
    return "MCP23008 IO Expander"
}

func (mcp *MCP23008) ReadRegister(register ChipRegister) (byte, error) {
    p := make([]byte, 1)
    out:=make([]byte,1  )
    out[0]=byte(register)
    err := mcp.conn.Tx(out, p)
	return p[0], err
}

func (mcp *MCP23008) ReadPullup(pin int) (bool, error) {
	cur, err := mcp.ReadRegister(GPPU)
	return (cur & (1 << pin)) == (1 << pin), err
}

func (mcp *MCP23008) SetPinMode(pin int, mode OutputMode) error {
	current, _ := mcp.ReadRegister(IODIR)
	newval := SetBit(current, pin, mode == ModeInput)
	var err error
	if newval != current {
		err = mcp.WriteRegister(IODIR, newval)
	}
	return err
}

func (mcp *MCP23008) SetPullup(pin int, state bool) error {
	current, err := mcp.ReadRegister(GPPU)
	new := SetBit(current, pin, state)
	if new == current {
		return nil
	} else {
		mcp.WriteRegister(GPPU, new)
	}
    return err
}

func (mcp *MCP23008) WriteGPIO(byteval byte) (int, error) {

	return 1, mcp.WriteRegister(GPIO, byteval)
}

func (mcp *MCP23008) WriteGPIOPin(pin int, on bool) (byte, error) {
	current, err := mcp.ReadGPIO()
	newval := SetBit(current, pin, on)
	if newval == current {
		return current, nil
	} else {
        _, err =mcp.WriteGPIO(newval)
        return current, err
	}
}

func (mcp *MCP23008) WriteRegister(register ChipRegister, value byte) error {
    buf := make([]byte, 2)
    buf[0]=byte(register)
    buf[1]=value
    in := make([]byte, 1)
	err := mcp.conn.Tx(buf, in)
    if err!=nil {
        log.Print(err)
    }
    return err
}

func SetBit(value byte, bit int, on bool) byte {
	if on {
		value |= maskOn[bit]
	} else {
		value &= maskOff[bit]
	}
	return value
}
