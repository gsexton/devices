package adafruit

import (
	"fmt"
	"log"
	"periph.io/x/conn"
	"periph.io/x/devices/lcd"
	"periph.io/x/devices/mcp23008"
	"time"
)

const (
	rsPin        = 1
	enablePin    = 2
	backlightPin = 7
)

type writeMode bool

const (
	modeCommand writeMode = false
	modeData    writeMode = true
)

var dataPins = []byte{3, 4, 5, 6}

type AdafruitI2CSPIBackpack struct {
	LCD *HD44780
	mcp *mcp23008.MCP23008
	on  bool
}

func NewAdafruitI2CSPIBackpack(conn conn.Conn, rows, cols int) *AdafruitI2CSPIBackpack {
	mcp, err := mcp23008.NewMCP23008(conn)
	if err != nil {
		log.Fatal(err)
	}
	bp := AdafruitI2CSPIBackpack{mcp: mcp}
	bp.LCD = NewHD44780(&bp, rows, cols)
	bp.init()
	return &bp
}

func (bp *AdafruitI2CSPIBackpack) String() string {
	return fmt.Sprintf("Adafruit I2C SPI Backpack %s, - Rows: %d, Cols: %d", bp.mcp.String(), bp.LCD.rows, bp.LCD.cols)
}

func (bp *AdafruitI2CSPIBackpack) init() {
	// Set the IO Direction for all pins to output.
	mcp := bp.mcp
	mcp.WriteRegister(mcp23008.IODIR, 0)
	/*
	   This is the startup sequence for the Hitachi HD44780U chip as
	   documented in the Datasheet. Note this is the init sequence for
	   4 bit mode.
	*/
	mcp.WriteGPIOPin(rsPin, bool(modeCommand))
	mcp.WriteGPIOPin(enablePin, false)

	bp.write4Bits(0x03)
	time.Sleep(4100 * time.Microsecond)

	bp.write4Bits(0x03)
	bp.write4Bits(0x03)
	bp.write4Bits(0x02)
	/*
	   function set 0x20
	       0x10 - 1 = 8 bit mode, 0= 4bit mode
	       0x08 - 1 = 2 lines, 0= 1 line.
	       0x04 - Font. 1=5*10 dots, 0=5*8 Dots
	       0x02 - Unused
	       0x01 - Unused
	*/
	var lineMode LCDCommand = 0x20
	if bp.LCD.rows > 1 {
		lineMode |= 0x08
	}
	bp.SendCommand([]LCDCommand{lineMode})
	/*
	   Display On/Off Control 0x08
	   0x04, 1=Display On, 0=Display Off
	   0x02, 1=Cursor On, 0=Cursor Off
	   0x01, 1=Blink On, 0= Blink Off
	*/
	bp.SendCommand([]LCDCommand{0x0c})
	bp.SendCommand([]LCDCommand{0x01})
	time.Sleep(3 * time.Millisecond)
	bp.SendCommand([]LCDCommand{0x02})
	
	bp.SetBacklight(1)
}

func (bp *AdafruitI2CSPIBackpack) write(value byte) error {
	_, err := bp.mcp.WriteGPIOPin(rsPin, bool(modeData))
	if err == nil {
		err = bp.write4Bits(value >> 4)
	}
	if err == nil {
		bp.write4Bits(value)
	}
	return err

}

func (bp *AdafruitI2CSPIBackpack) Write(p []byte) (n int, err error) {
	// log.Print("Write(",p,")")
	if len(p)==0 {
		return 0, nil
	}
	if p[0]==cmdByte {
		bp.SendCommand([]LCDCommand{LCDCommand(p[1])})
		return
	}
	n = len(p)
	_, err = bp.mcp.WriteGPIOPin(rsPin, bool(modeData))
	if err != nil {
		log.Print(err)
		return
	}
	for _, byteVal := range p {
		err = bp.write4Bits(byteVal >> 4)
		err = bp.write4Bits(byteVal & 0x0f)
		if err != nil {
			break
		}
	}
	if err != nil {
		log.Print(err)
	}
	return
}

func (bp *AdafruitI2CSPIBackpack) SetBacklight(intensity lcd.Intensity) {
	bp.on = (intensity > 0)
	// log.Print("SetBacklight(",intensity,",) on = ",bp.on)
	bp.LCD.SetDisplay(bp.on)
	_, err := bp.mcp.WriteGPIOPin(backlightPin, bp.on)
	if err != nil {
		log.Fatal(err)
	}
}

func (bp *AdafruitI2CSPIBackpack) SendCommand(commands []LCDCommand) error {
	var err error
	_, err = bp.mcp.WriteGPIOPin(rsPin, bool(modeCommand))
	if err != nil {
		log.Print(err)
	}
	for _, command := range commands {
		bp.write4Bits(byte(command >> 4))
		bp.write4Bits(byte(command))
	}
	return err
}

func (bp *AdafruitI2CSPIBackpack) write4Bits(value byte) error {
	writeVal, err := bp.mcp.ReadGPIO()
	value = value & 0x0f
	// log.Print("write4Bits(",strconv.FormatInt(int64(value),2),") current ioport value: ",strconv.FormatInt(int64(writeVal),2))
	writeVal &= 0x83
	writeVal|=(value << 3)
	
	if err != nil {
		return err
	}
	// log.Print("write4Bits(value=(",value,") ",strconv.FormatInt(int64(value),2),") Writing to GPIO: ",strconv.FormatInt(int64(writeVal),2))
	
	_, err = bp.mcp.WriteGPIO(writeVal)

	writeVal|=0x04
	_, err = bp.mcp.WriteGPIO(writeVal)
	
	writeVal &= 0xfb
	_, err = bp.mcp.WriteGPIO(writeVal)

	// writeVal, err = bp.mcp.ReadGPIO()
	// log.Print("re-reading value returns: ",strconv.FormatInt(int64(writeVal),2))

	return err
}
