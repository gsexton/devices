// Copyright 2025 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

// This package provides a driver for the Microchip MCP472x Series of Digital
// to Analog converters. It works with the MCP4725 and MCP4728 chips. The
// MCP4728 shares a common command interface, but offers 4 outputs and has
// a precision internal voltage reference. The MCP4725 uses VCC for vRef.
//
// # Datasheets
//
// # MCP4725
//
// https://ww1.microchip.com/downloads/en/devicedoc/22039d.pdf
//
// # MCP4728
//
// https://www.digikey.com/htmldatasheets/production/623709/0/0/1/mcp4728.html
package mcp472x

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"periph.io/x/conn/v3/i2c"
	"periph.io/x/conn/v3/physic"
)

// Variant represents the model of the device.
type Variant string

const (
	// The internal precision reference for the MCP4728
	MCP4728_INTERNAL_REF physic.ElectricPotential = 2048 * physic.MilliVolt
	VARIANT_MCP4725      Variant                  = "MCP4725"
	VARIANT_MCP4728      Variant                  = "MCP4728"
	stepCount                                     = 1 << 12 // 12 bit a/d
	maxCount                                      = stepCount - 1

	errInvalidVoltage             = "mcp472x: voltage %s out of range"
	errBusy                       = "mcp472x: device busy"
	errCountOutOfRange            = "mcp472x: count out of range"
	errInvalidInputCount          = "mcp472x: invalid number of inputs provided"
	DefaultAddress       i2c.Addr = 0x60
	Channels4728                  = 4
	Channels4725                  = 1
)

// Channel PowerDown mode.
type PDMode byte

const (
	PDModeNormal PDMode = iota
	// The remaining values specify resistance value used to tie the output pin
	// to ground.
	PDMode_1K
	PDMode_100K
	PDMode_500K
)

// Dev represents an MCP472X DtoA converter.
type Dev struct {
	d           i2c.Dev
	variant     Variant
	maxChannels int
	vRef        physic.ElectricPotential
}

// SetOutputParam represents a parameter for programming a DAC output
// channel.
type SetOutputParam struct {
	// DAC is the D/A channel number. Always 0 for MCP4725.
	DAC byte
	// V is the voltage to set the output to. For external reference, the
	// value can be 0-VCC. For the MCP4728 using the internal precision
	// reference, the value can be 0 - (2 * MCP4728_INTERNAL_REF)
	V physic.ElectricPotential
	// For the MCP4728 using the internal reference, you can boost the
	// gain of the output to 2 * MCP4728_INTERNAL_REF, at the cost of
	// resolution. VOut cannot exceed VCC. If you need vOut > 3.3V,
	// ensure VCC=5v.
	BoostGain bool
	// True to use the internal reference. Used only for MCP4728.
	UseInternalRef bool
	// Powerdown mode for this channel
	PDMode PDMode
}

// New creates and returns a representation of a MCP472X Digital to Analog
// converter. vRef sets the reference voltage used by the device. The value of
// vRef is used to convert a voltage parameter to a count value that the
// device internally uses for settings.
//
// For the MCP4728, the internal 2.048v reference can be used, or you can
// specify that VCC be used. For the internal reference, you can use the default
// gain which will have full scale voltage be 0-2.048v, or set gain to 2, which
// will make output be from 0-4.096v. Note that VCC must be 5V in this case.
//
// For the MCP4725, VCC is used as vRef.
func New(bus i2c.Bus, addr i2c.Addr, variant Variant, vRef physic.ElectricPotential) (*Dev, error) {

	d := &Dev{
		d:           i2c.Dev{Bus: bus, Addr: uint16(addr)},
		variant:     variant,
		vRef:        vRef,
		maxChannels: Channels4728,
	}
	if variant == VARIANT_MCP4725 {
		d.maxChannels = Channels4725
	}
	return d, nil
}

// PotentialToCount converts the specified voltage to the count for the
// A/D converter. It returns the required count, whether boostGain should
// be enabled. If the voltage is negative, or otherwise out of range, an
// error is returned. The count will roughly be v/(vRef/4095).
func (d *Dev) PotentialToCount(v physic.ElectricPotential) (uint16, bool, error) {
	if (v < 0) || (v > d.vRef && d.variant == VARIANT_MCP4725) || (v > (2 * d.vRef)) {
		return 0, false, fmt.Errorf(errInvalidVoltage, v)
	}
	boost := false
	stepValue := d.vRef / maxCount
	count := uint16(float64(v)/float64(stepValue) + 0.5)
	if count > maxCount && d.variant == VARIANT_MCP4728 {
		boost = true
		count = count >> 1
	}
	return count, boost, nil
}

// Convert the current and eeprom registers for a 4725. The bit structure is
// different between the two models...
func (d *Dev) convert4725(bytes []byte) ([]SetOutputParam, []SetOutputParam, error) {
	busy := bytes[0]&0x80 == 0x0
	step := float64(d.vRef) / float64(maxCount)
	count := float64((uint16(bytes[1]&0xf0) << 4) | (uint16(bytes[1]&0x0f) << 4) | (uint16(bytes[2]) >> 4))
	op := SetOutputParam{V: physic.ElectricPotential(step * count)}
	op.PDMode = PDMode((bytes[0] >> 1) & 0x03)
	current := []SetOutputParam{op}

	count = float64(uint16(bytes[4]) | (uint16(bytes[3]&0x0f) << 8))
	op = SetOutputParam{V: physic.ElectricPotential(step * count)}
	op.PDMode = PDMode((bytes[3] >> 5) & 0x03)
	eeprom := []SetOutputParam{op}
	var err error
	if busy {
		err = errors.New(errBusy)
	}
	return current, eeprom, err
}

// convert4728 converts the bytes read for current and EEPROM registers for
// an MCP4728 to SetOutputParams Refer to the datasheet for more information.
func (d *Dev) convert4728(bytes []byte) ([]SetOutputParam, []SetOutputParam, error) {
	step := d.vRef / maxCount
	current, eeprom := make([]SetOutputParam, 0), make([]SetOutputParam, 0)
	pos := 0
	vals := make([]SetOutputParam, 2)
	busy := false
	// A-D
	for channelID := range d.maxChannels {
		// Current Output Parameters, and EEPROM Parameters
		for i := range 2 {
			busy = busy || bytes[pos]&0x80 == 0x0
			count := physic.ElectricPotential(uint16(bytes[pos+2]) | uint16(bytes[pos+1]&0x0f)<<8)
			pdMode := PDMode(bytes[pos+1] >> 5 & 0x03)
			boost := bytes[pos+1]&0x10 == 0x10
			vref := bytes[pos+1]&0x80 == 0x80
			if boost && vref {
				// Boost Gain is turned on, and it's using the internal reference, so
				// double the count...
				count *= 2
			}
			vals[i] = SetOutputParam{
				DAC:            byte(channelID),
				V:              physic.ElectricPotential(step * count),
				PDMode:         pdMode,
				BoostGain:      boost,
				UseInternalRef: vref,
			}
			pos += 3
		}
		current = append(current, vals[0])
		eeprom = append(eeprom, vals[1])
	}
	var err error
	if busy {
		err = errors.New(errBusy)
	}
	return current, eeprom, err
}

// FastWrite sends raw AD count values to the converter. Bits 0-11 represent,
// the count, and bits 12 and 13 the PowerDown mode, if applicable. Use
// PotentialToCount to convert a specific voltage value to the count value.
func (d *Dev) FastWrite(values ...uint16) (err error) {
	if len(values) != d.maxChannels {
		err = errors.New(errInvalidInputCount)
		return
	}
	w := make([]byte, len(values)*2)
	for ix, val := range values {
		w[ix*2] = byte(val>>8) & 0x3f // mask off the two high bits.
		w[ix*2+1] = byte(val & 0xff)
	}
	err = d.d.Tx(w, nil)
	return
}

// GetOutput reads the configured output values and programmed EEPROM values,
// and returns the results. If the device signals it is busy with an EEPROM
// write, the function will retry up to 9 times to read values.
func (d *Dev) GetOutput() (current []SetOutputParam, eeprom []SetOutputParam, err error) {
	var r []byte
	bytesChannel := 5
	if d.variant == VARIANT_MCP4728 {
		bytesChannel = 6
	}
	r = make([]byte, bytesChannel*d.maxChannels)
	for range 10 {
		err = d.d.Tx(nil, r)
		if err != nil {
			return
		}
		if d.variant == VARIANT_MCP4725 {
			current, eeprom, err = d.convert4725(r)
		} else {
			current, eeprom, err = d.convert4728(r)
		}
		if err == nil {
			break
		}
		// The device is busy with an eeprom write. Wait and try again.
		time.Sleep(100 * time.Millisecond)
	}
	return
}

// SetOutput sets the output of the specified output channels. For an MCP4725, you may
// pass only one param.
func (d *Dev) SetOutput(params ...SetOutputParam) (err error) {
	if len(params) == 0 || len(params) > d.maxChannels {
		err = errors.New(errInvalidInputCount)
		return
	}

	if d.variant == VARIANT_MCP4725 {
		w := d.paramToBytes(&params[0])
		err = d.d.Tx(w, nil)
		return
	}
	var w []byte
	w = make([]byte, 0)
	for _, param := range params {
		bytes := d.paramToBytes(&param)
		w = append(w, bytes...)
	}
	w[0] |= 0x40
	err = d.d.Tx(w, nil)
	return
}

// SetOutputWithSave sets the channel output values AND saves the values to
// EEPROM. On powerup, the output value will be set to the saved settings.
func (d *Dev) SetOutputWithSave(params ...SetOutputParam) (err error) {
	if len(params) == 0 || len(params) > d.maxChannels {
		err = errors.New(errInvalidInputCount)
		return
	}
	var w []byte
	for _, param := range params {
		bytes := d.paramToBytes(&param)
		w = append(w, bytes...)
	}

	if d.variant == VARIANT_MCP4725 {
		w[0] |= 0x60
		err = d.d.Tx(w, nil)
		return
	}

	w[0] |= 0x50
	if len(params) == 1 {
		w[0] |= 0x08
	}

	err = d.d.Tx(w, nil)
	return

}

// Implements conn.Resource
func (d *Dev) Halt() error {
	return nil
}

func (d *Dev) String() string {
	return string(d.variant)
}

// paramToBytes converts a SetOutputParam to the appropriate bit structure for
// write. The bit structure for writes is different depending on the variant.
func (d *Dev) paramToBytes(op *SetOutputParam) (bytes []byte) {
	count, boost, _ := d.PotentialToCount(op.V)
	if d.variant == VARIANT_MCP4725 {
		b := byte(0b01000000)
		b |= (byte(op.PDMode&0x03) << 1)
		bytes = append(bytes, b)
		bytes = append(bytes, byte(count>>4)&0xff)
		bytes = append(bytes, byte(count<<4)&0xf0)
		return
	}
	// 4728
	// set the DAC Channel bits, and UDAC
	bytes = append(bytes, byte(((op.DAC&0x03)<<1)&0xff))
	b := byte(count>>8) & 0xff
	if op.UseInternalRef {
		b |= 0x80
	}
	if boost {
		b |= 0x10
	}

	b |= byte(op.PDMode&0x03) << 5
	bytes = append(bytes, b)
	bytes = append(bytes, byte(count&0xff))
	return
}

// Equal compares two SetOutputParam values for equality.
func (op SetOutputParam) Equal(op2 SetOutputParam) bool {
	return op.V == op2.V &&
		op.BoostGain == op2.BoostGain &&
		op.UseInternalRef == op2.UseInternalRef &&
		op.PDMode == op2.PDMode
}

func (op SetOutputParam) String() string {
	bytes, _ := json.Marshal(&op)
	return string(bytes)
}
