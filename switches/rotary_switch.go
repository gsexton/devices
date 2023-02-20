package switches

import (
	"log"
	"periph.io/x/conn/gpio"
)

type RotarySwitchValue int

const (
	DirectionCCW RotarySwitchValue = iota
	DirectionCW
	ButtonPress
	ButtonRelease
)

const stateInvalid=0xf

// A rotary encoder like the Adafruit Rotary Switch/Push Button Switch Combo.
//
// https://www.adafruit.com/product/377
//
type RotarySwitch struct {
	ch                           chan RotarySwitchValue
    // Number of switch events that should be buffered. If 
    // the channel becomes full, then events will be dropped.
	channelSize                  int
	position                     int
    // If true, and the position exceeds the max_positions or less than zero,
    // wrap appropriately.
    wrap_position                bool
    // The maximum number of logical positions the switch should have.
	max_positions                int
	st_pin, data_pin, button_pin gpio.PinIO
    // Internal variable for tracking switch movement and debounce.
	state                        int
	last_button                  gpio.Level
    // If true, widen up the allowed values for transition to work around
    // inherent issues of rotary switches.
    loose                        bool
	terminating                  bool
}

// Retrieve the channel that switch events will be
// sent to.
func (sw *RotarySwitch) Channel() chan RotarySwitchValue {
	return sw.ch
}

// Rotary switches can be difficult. Setting loose mode = true
// relaxes the test because rotary switches are bouncy, and the
// time to read the data line might take so long that the 
// data line transitions to the next state. For the Adafruit
// switch, this seems to be really common when going in the
// clockwise direction.
func (sw *RotarySwitch) SetLooseMode(mode bool) {
    sw.loose = mode
}

// Default true. If wrap is set, and the switch
// position exceeds the max positions, then the
// value for position will be wrapped to 0. Similarly, if the 
// position becomes less than 0, then it will wrap to
// max position. If mode = false, then position will limit to
// zero and max position.
func (sw *RotarySwitch) SetWrap(mode bool) {
    sw.wrap_position = mode
}

// Close the IO Lines associated with this switch.
func (sw *RotarySwitch) Close() {
	if sw.st_pin == nil {
		return
	}
	sw.terminating = true
	sw.st_pin.In(gpio.PullNoChange, gpio.NoEdge)
	sw.button_pin.In(gpio.PullNoChange, gpio.NoEdge)
	close(sw.ch)
	sw.st_pin = nil
	sw.data_pin = nil
	sw.button_pin = nil
}

// You can read the relative position using Position() and the returned value
// will be be between zero and maxPositions-1
//
// When the switch is turned clockwise, the position is incremented, and
// when the switch is turned counter-clockwise, the position is decremented.
//
// If wrap is set, and the bounds are exceeded, the position is set to the 
// opposite bound.
//
func (sw *RotarySwitch) Position() int {
	return sw.position
}


// The goroutine that listens for edge changes on the rotary switch.
func rotaryHandler(sw *RotarySwitch) {
	for {
		sw.st_pin.WaitForEdge(-1)
		if sw.terminating {
			return
		}
        sw.state = (sw.state & 0x03) << 2
        if sw.data_pin.Read() == gpio.High {
            sw.state |= 0x01
        }
		if sw.st_pin.Read() == gpio.High {
			sw.state |= 0x02
		}
		/*
				 * This code listens for the raw pin events on the digital io
				 * pins and processes them into Switch rotation events.
				 *
				 * A single turn of the rotary encoder generates two events on
				 * the State pin, a Low Transition, followed by a High
				 * transition. CW versus CCW is done by examining the Data pin.
				 * The transition table for turns is:
				 *
				 *  CW
				 *    State: Low    Data: High
		         *    State: High   Data: Low
		         *
		         *    1101 0110
				 *
		         *  CCW
		         *    1100 0011
		         *
				 *    State: Low    Data: Low
				 *    State: High   Data: High
		*/
        /*
        log.Printf("sw.state=%x",sw.state)
		if last == (sw.state & 0x03) {
            log.Print("Rotary Switch bounce event cancelled.")
			continue
		}
        */

		if len(sw.ch) == sw.channelSize {
			// If the channel is full, just stop.
			log.Print("RotarySwitch event channel full. Dropping event.")
			continue
		}
        // This is nuts. According to design, this should be 0x06. The only thing I can
        // think is that by time the read happens, the data line has already changed
        // state. This is true even when using a low latency controller like a Pi Pico
        // with the Arduino C binaries...
		if sw.state == 0x06 || (sw.loose && sw.state==0x07) {
			sw.position = sw.position + 1
			if sw.position >= sw.max_positions {
                if sw.wrap_position {
                    sw.position = 0
                } else {
                    sw.position = sw.max_positions
                }
			}
			sw.ch <- DirectionCW
			sw.state = stateInvalid
		} else if sw.state == 0x03 {
			sw.position = sw.position - 1
			if sw.position < 0 {
                if sw.wrap_position {
                    sw.position = sw.max_positions - 1
                } else {
                    sw.position = 0
                }
			}
			sw.ch <- DirectionCCW
			sw.state = stateInvalid
		}
	}
}

// The go routine that listens for button pushes on the
// tactile button switch.
func buttonHandler(sw *RotarySwitch) {
	for {
		sw.button_pin.WaitForEdge(-1)
		if sw.terminating {
			return
		}
		state := sw.button_pin.Read()
		if len(sw.ch) < sw.channelSize {
			if state == gpio.High && sw.last_button == gpio.Low {
				sw.ch <- ButtonRelease
			} else if state == gpio.Low && sw.last_button == gpio.High {
				sw.ch <- ButtonPress
			}
			sw.last_button = state
		}

	}
}

// Construct a new Rotary Switch by passing in the gpio.PinIO.
//
// maxPosition can be used to  keep track of the relative position of the switch.
//
// To receive events, read the Channel returned by RotarySwitch.Channel(). The
// value returned will be one of the constants.
//
// If you're using a Rotary Switch that doesn't include a button switch, pass
// gpio.INVALID for buttonPin.
func NewRotarySwitch(statePin, dataPin, buttonPin gpio.PinIO, maxPositions int) *RotarySwitch {
	statePin.In(gpio.PullUp, gpio.BothEdges)
	dataPin.In(gpio.PullUp, gpio.NoEdge)

	channelSize := 16
	sw := RotarySwitch{ch: make(chan RotarySwitchValue, channelSize),
		st_pin:        statePin,
		data_pin:      dataPin,
		button_pin:    buttonPin,
		max_positions: maxPositions,
        wrap_position: true,
		channelSize:   channelSize,
		state:         stateInvalid,
		last_button:   gpio.Low,
        loose:         true}

	go rotaryHandler(&sw)
	if buttonPin != gpio.INVALID {
		buttonPin.In(gpio.PullUp, gpio.BothEdges)
		go buttonHandler(&sw)
	}

	log.Print("Rotary switch provisioned.")
	return &sw
}
