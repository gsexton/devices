package lcd

import (
    "periph.io/x/conn"
    "sync"
)

type CursorDirection int
const (
    // Move the cursor one unit back.
    Backward CursorDirection = iota
    // Move the cursor one unit forward.
    Forward
    Up
    Down
)

type CursorMode int
const (
    // Turn the cursor Off
	CursorOff CursorMode = iota
    // Enable Underline Cursor
	CursorUnderline
    // Enable Block Cursor
	CursorBlock
    // Blinking
    CursorBlink
)

type LCDPort struct {
    Conn conn.Conn
    Mu sync.Mutex
}

type LCD interface {
    // Enable/Disble auto scroll
    AutoScroll(enabled bool)
    // Return the number of columns the LCD Supports
	Cols() int
    // Clear the display and move the cursor home.
	Clear()
    // Set the cursor mode. You can pass multiple arguments.
    // Cursor(CursorOff, CursorUnderline)
    Cursor(mode ...CursorMode)
    // Move the cursor home (1,1)
	Home()
    // Return the min column position.
    MinCol() int
    // Return the min row position.
    MinRow() int
    // Move the cursor forward or backward.
	Move(dir CursorDirection)
    // Move the cursor to arbitrary position (1,1 base).
	MoveTo(row, col int)
    // Return the number of rows the LCD supports.
    Rows() int
    // Turn the display on / off
    SetDisplay(on bool)
    // Write a set of bytes to the display.
    Write(p []byte) (n int, err error)
    // Write a string output to the display.
    WriteString(text string) (n int, err error)


}

type Intensity int
// Interface for displays that support a monochrome Backlight.
// Displays that support RGB intensity should implement this
// as well.
//
// As a side note, a lot of the units that support this command
// write the value to eeprom, which has a finite number of
// writes. To turn the unit on/off, use LCD.SetDisplay()
type Backlight interface {
    SetBacklight(intensity Intensity)
}

// Interface for displays that support a RGB Backlight.
type RGBBacklight interface {
    SetBacklight(red, green, blue Intensity)
}

type Contrast int
// Interface for displays that support a contrast adjust
// function.
type LCDContrast interface {
    SetContrast(contrast Contrast)
}
