package adafruit

import (
	"errors"
	"fmt"
	"io"
	"log"
	iolcd "periph.io/x/devices/v3/lcd"
	"sync"
    "time"
)

type HD44780 struct {
	w    io.Writer
	mu   sync.Mutex
	rows int
	cols int
	on   bool
    cursor bool
    blink bool
}

type LCDCommand byte

const cmdByte byte = 0xfe

type commandID int

const (
	autoScrollOff commandID = iota
	autoScrollOn
	blockCursorOff
	blockCursorOn
	clearScreen
	cursorBack
	cursorForward
	displayOff
	displayOn
	goHome
	setCursorPosition
	underlineCursorOff
	underlineCursorOn
)

var rowConstants = [][]byte{{0, 0, 64}, {0, 0, 64, 20, 84}}

func getCommand(commandNumber commandID) []byte {
	var result []byte
	switch commandNumber {
	case autoScrollOff:
		result = []byte{cmdByte, 0x06}
	case autoScrollOn:
		result = []byte{cmdByte, 0x07}
	case clearScreen:
		result = []byte{cmdByte, 0x01}
	case cursorBack:
		result = []byte{cmdByte, 0x10}
	case cursorForward:
		result = []byte{cmdByte, 0x14}
	case displayOff:
		result = []byte{cmdByte, 0x08}
	case displayOn:
		result = []byte{cmdByte, 0x0c}
	case goHome:
		result = []byte{cmdByte, 0x02}
	case setCursorPosition:
		result = []byte{cmdByte, 0x80}
	case underlineCursorOff:
		result = []byte{cmdByte, 0x4b}
	case underlineCursorOn:
		result = []byte{cmdByte, 0x0e}
	default:
		log.Fatal(errors.New("Unhandled command id"))
	}
	return result
}

func getRowConstant(row, maxcols int) byte {
	var offset int
	if maxcols != 16 {
		offset = 1
	}
	return rowConstants[offset][row]
}
func NewHD44780(io io.Writer, rows, cols int) *HD44780 {
	display := HD44780{w: io, rows: rows, cols: cols, on: true}
	return &display
}
func (lcd *HD44780) AutoScroll(enabled bool) {
	if enabled {
		lcd.w.Write(getCommand(autoScrollOn))
	} else {
		lcd.w.Write(getCommand(autoScrollOff))
	}
}


func (lcd *HD44780) Clear() {
	lcd.w.Write(getCommand(clearScreen))
    time.Sleep(2 * time.Millisecond)
}

func (lcd *HD44780) Cols() int {
	return lcd.cols
}

func (lcd *HD44780) Cursor(modes ...iolcd.CursorMode) {
	var val = byte(0x08)
	if lcd.on {
		val |= 0x04
	}
	for _, mode := range modes {
		switch mode {
		case iolcd.CursorOff:
			lcd.w.Write(getCommand(underlineCursorOff))
            lcd.blink=false
            lcd.cursor=false
        case iolcd.CursorBlink:
            lcd.blink=true
            lcd.cursor=true
			val |= 0x03
        case iolcd.CursorUnderline:
            lcd.cursor=true
            lcd.blink=true
			lcd.w.Write(getCommand(underlineCursorOn))
        case iolcd.CursorBlock:
            lcd.cursor=true
            lcd.blink=true
			val |= 0x07
		}
	}
	lcd.w.Write([]byte{cmdByte, val})
}

func (lcd *HD44780) Home() {
	lcd.w.Write(getCommand(goHome))
    time.Sleep(2 * time.Millisecond)
}

func (lcd *HD44780) MinCol() int {
	return 1
}

func (lcd *HD44780) MinRow() int {
	return 1
}

func (lcd *HD44780) Move(direction iolcd.CursorDirection) {
	if direction == iolcd.Forward {
		lcd.w.Write(getCommand(cursorForward))
	} else {
		lcd.w.Write(getCommand(cursorBack))
	}
}

func (lcd *HD44780) MoveTo(row, col int) {
	if row < lcd.MinRow() || row > lcd.rows || col < lcd.MinCol() || col > lcd.cols {
		log.Print("HD44780.MoveTo(", row, ",", col, ") value out of range.")
		return
	}
	cmd := getCommand(setCursorPosition)
	cmd[1] |= getRowConstant(row, lcd.cols) + byte(col-1)
	lcd.w.Write(cmd)
}
func (lcd *HD44780) Rows() int {
	return lcd.rows
}
func (lcd *HD44780) String() string {
	return fmt.Sprintf("Adafruit USB LCD Backpack. Rows: %d Cols: %d", lcd.rows, lcd.cols)
}

func (lcd *HD44780) SetDisplay(on bool) {
	lcd.on = on
    val:=byte(0x08)
    if on {
        val|=0x04
    }
    if lcd.blink {
        val|=0x01
    }
    if lcd.cursor {
        val|=0x02
    }
	lcd.w.Write([]byte{cmdByte, val})
}

func (lcd *HD44780) Write(p []byte) (n int, err error) {
    if len(p) == 0 {
        return 0, nil
    }
	lcd.mu.Lock()
	defer lcd.mu.Unlock()
	return lcd.w.Write(p)
}

func (lcd *HD44780) WriteString(text string) (int, error) {
    lcd.mu.Lock()
	defer lcd.mu.Unlock()
	return lcd.w.Write([]byte(text))
}


