package main

import (
	"machine"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"
)

// Currently still an issue with atomics in TinyGo, so use
// a mutex to protect our variables here.
type Sensor struct {
	interval int // in seconds
	stop     bool
	lock     sync.RWMutex
}

func main() {
	// Initialize the sensor settings.
	s := Sensor{
		interval: 1,
		stop:     false,
	}

	// Configure the serial port.
	machine.Serial.Configure(machine.UARTConfig{
		BaudRate: 115200,
	})

	// Start a goroutine to continuously read serial data.
	go s.receiveSerial()

	// Main loop: read and print temperature periodically.
	for {
		// Get the interval and stop values. It's ok if we are off by a loop iteration.
		s.lock.RLock()
		interval := s.interval
		stop := s.stop
		s.lock.RUnlock()

		if stop {
			runtime.Gosched()
			continue
		}

		ReadTemperature()
		time.Sleep(time.Duration(interval) * time.Second)
	}
}

// ReadTemperature reads and converts the raw temperature reading.
// Also sends data on serial port.
func ReadTemperature() {
	temp := float32(machine.ReadTemperature()) / 1000
	// Send the temperature data over serial.
	// Note: println in tinygo will send to serial by default.
	println(temp)
}

// receiveSerial continuously reads incoming data from the serial port.
func (s *Sensor) receiveSerial() {

	for {
		// Read from the serial interface.
		// If there's data in the Rx buffer, call ReadByte to read it until it's empty.
		n := machine.Serial.Buffered()

		if n > 0 {
			// Process the received data.
			buf := make([]byte, n)
			for i := 0; i < n; i++ {
				b, _ := machine.Serial.ReadByte()
				buf[i] = b
			}

			// Allowed commands
			// interval X: Set the read (and send) interval to X seconds.
			// stop: Stop sending temperature data.
			// start: Start sending temperature data.
			// Note: for simplicity, for bad commands or malformed ones, we just ignore them.
			command := strings.Fields(string(buf))
			if len(command) > 0 {
				switch command[0] {
				case "interval":
					if len(command) > 1 {
						// Parse the interval value.
						s.lock.Lock()
						interval, err := strconv.Atoi(command[1])
						if err == nil {
							// Set the interval to the new value.
							s.interval = interval
						}
						s.lock.Unlock()
					}
				case "stop":
					// Stop sending temperature data.
					s.lock.Lock()
					s.stop = true
					s.lock.Unlock()
				case "start":
					// Start sending temperature data.
					s.lock.Lock()
					s.stop = false
					s.lock.Unlock()
				}
			}
		}

		// Yield to other goroutines. (Needed in TinyGo to allow other goroutines to run.)
		runtime.Gosched()
	}
}
