// Copyright 2025 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package mcp472x_test

import (
	"log"
	"time"

	"periph.io/x/conn/v3/i2c/i2creg"
	"periph.io/x/conn/v3/physic"
	"periph.io/x/devices/v3/mcp472x"
	"periph.io/x/host/v3"
)

// Simple Demo output voltage.
func Example() {
	if _, err := host.Init(); err != nil {
		log.Fatal("Error calling host.init()")
	}
	bus, err := i2creg.Open("")
	if err != nil {
		log.Fatal(err)
	}
	// To use VCC for the reference voltage, specify it as: 3_300 * physic.MilliVolt, etc.
	dev, err := mcp472x.New(bus, mcp472x.DefaultAddress, mcp472x.VARIANT_MCP4728, mcp472x.MCP4728_INTERNAL_REF)
	if err != nil {
		log.Fatal(err)
	}

	// Program channel 3 to output 512mV
	op := mcp472x.SetOutputParam{DAC: 3, V: 512 * physic.MilliVolt, UseInternalRef: true}
	err = dev.SetOutput(op)
	if err != nil {
		log.Println(err)
	}
	time.Sleep(10 * time.Second)

	// Power the channel down
	op.V = 0
	op.PDMode = mcp472x.PDMode_1K
	_ = dev.SetOutput(op)
}
