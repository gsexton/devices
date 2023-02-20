package adafruit

import (
	"errors"
	"fmt"
	"io"
	"log"
	"periph.io/x/conn"
	iolcd "periph.io/x/devices/v3/lcd"
)

type USBLCDBackpack struct {
	iolcd.LCDPort
	rows      int
	cols      int
	contrast  iolcd.Contrast
	backlight iolcd.Intensity
}

func usbCommand(commandNumber commandID) []byte {
	var result []byte
	switch commandNumber {
	case autoScrollOff:
		result = []byte{cmdByte, 0x52}
	case autoScrollOn:
		result = []byte{cmdByte, 0x51}
	case blockCursorOff:
		result = []byte{cmdByte, 0x54}
	case blockCursorOn:
		result = []byte{cmdByte, 0x53}
	case clearScreen:
		result = []byte{cmdByte, 0x58}
	case cursorBack:
		result = []byte{cmdByte, 0x4c}
	case cursorForward:
		result = []byte{cmdByte, 0x4d}
	case displayOff:
		result = []byte{cmdByte, 0x46}
	case displayOn:
		result = []byte{cmdByte, 0x42}
	case goHome:
		result = []byte{cmdByte, 0x48}
	case setCursorPosition:
		result = []byte{cmdByte, 0x47}
	case underlineCursorOff:
		result = []byte{cmdByte, 0x4b}
	case underlineCursorOn:
		result = []byte{cmdByte, 0x4a}
	default:
		log.Fatal(errors.New("Unhandled command id"))
	}
	return result
}

func NewUSBLCDBackpack(conn conn.Conn, rows, cols int) *USBLCDBackpack {
	bp := USBLCDBackpack{LCDPort: iolcd.LCDPort{Conn: conn}, rows: rows, cols: cols}
	return &bp
}

func (lcd *USBLCDBackpack) String() string {
	return fmt.Sprintf("Adafruit USB LCD Backpack. Rows: %d Cols: %d Connection: %s", lcd.rows, lcd.cols, lcd.Conn.String())
}

func (lcd *USBLCDBackpack) Cols() int {
	return lcd.cols
}

func (lcd *USBLCDBackpack) Rows() int {
	return lcd.rows
}

func (lcd *USBLCDBackpack) Clear() {
	lcd.sendCommand(clearScreen)
}

func (lcd *USBLCDBackpack) Home() {
	lcd.MoveTo(1, 1)
}

func (lcd *USBLCDBackpack) MoveTo(row, col int) {
	if row < 1 || row > lcd.rows || col < 1 || col > lcd.cols {
		log.Print("USBLCDBackpack.MoveTo(", row, ",", col, ") value out of range.")
		return
	}
	lcd.sendCommand(setCursorPosition, byte(row), byte(col))
}

func (lcd *USBLCDBackpack) Move(direction iolcd.CursorDirection) {
	if direction == iolcd.Forward {
		lcd.sendCommand(cursorForward)
	} else {
		lcd.sendCommand(cursorBack)
	}
}

func (lcd *USBLCDBackpack) AutoScroll(enabled bool) {
	if enabled {
		lcd.sendCommand(autoScrollOn)
	} else {
		lcd.sendCommand(autoScrollOff)
	}
}

func (lcd *USBLCDBackpack) Cursor(modes ...iolcd.CursorMode) {
	for _, mode := range modes {
		switch mode {
		case iolcd.CursorOff:
			lcd.sendCommand(blockCursorOff)
			lcd.sendCommand(underlineCursorOff)
		case iolcd.CursorUnderline:
			lcd.sendCommand(underlineCursorOn)
		case iolcd.CursorBlock:
			lcd.sendCommand(blockCursorOn)
		}
	}
}

func (lcd *USBLCDBackpack) setGPO(pin int, on bool) {
	if on {
		lcd.sendCommand(0x57, byte(pin))
	} else {
		lcd.sendCommand(0x56, byte(pin))
	}
}

func (lcd *USBLCDBackpack) sendCommand(command commandID, additional ...byte) {
	cmd := usbCommand(command)
	for _, cmdByte := range additional {
		cmd = append(cmd, cmdByte)
	}
	bytes := make([]byte, len(cmd))
	for ix, val := range cmd {
		bytes[ix] = byte(val)
	}
	lcd.Write(bytes)
}

func (lcd *USBLCDBackpack) Write(p []byte) (n int, err error) {
	lcd.Mu.Lock()
	defer lcd.Mu.Unlock()
	err = lcd.Conn.Tx(p, nil)
	n = len(p)
	return n, err
}

func (lcd *USBLCDBackpack) WriteString(text string) (int, error) {
	return lcd.Write([]byte(text))
}

func (lcd *USBLCDBackpack) Close() error {
	var cl io.Closer
	cl, ok := lcd.Conn.(io.Closer)
	var err error
	if ok {
		err = cl.Close()
	} else {
		err = errors.New("Conn (" + lcd.Conn.String() + ") doesn't implement io.Closer interface.")
	}

	return err
}

func (lcd *USBLCDBackpack) SetDisplay(on bool) {
	if on {
		lcd.sendCommand(displayOn)
	} else {
		lcd.sendCommand(displayOff)
	}
}

func (lcd *USBLCDBackpack) SetBacklight(intensity iolcd.Intensity) {
	lcd.sendCommand(0x99, byte(intensity))
}

func (lcd *USBLCDBackpack) SetContrast(contrast iolcd.Contrast) {
	lcd.sendCommand(0x50, byte(contrast))
}
