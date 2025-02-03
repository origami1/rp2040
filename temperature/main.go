package main

import (
	"image/color"
	"machine"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	"tinygo.org/x/drivers/ws2812"
)

const (
	minInterval    = 100 // Minimum interval in milliseconds.
	sendingLedTime = 50  // Time in milliseconds to keep the LED in the sending color.
)

var (
	// The onboard LED is used to indicate that the sensor is running.
	// It is a bit particular: set the color to green first, then you can set any color.
	neo machine.Pin = 25

	// Colors for the NeoPixel.
	black   = color.RGBA{R: 0x00, G: 0x00, B: 0x00}
	red     = color.RGBA{R: 0xff, G: 0x00, B: 0x00}
	green   = color.RGBA{R: 0x00, G: 0xff, B: 0x00}
	yellow  = color.RGBA{R: 0x00, G: 0xff, B: 0xff}
	off     = red
	on      = green
	sending = black // Setting to black is more visible.
)

// Currently still an issue with atomics in TinyGo, so use
// a mutex to protect our variables here.
type SensorSettings struct {
	interval  int // in milliseconds
	stop      bool
	ledDriver *ws2812.Device
	lock      sync.RWMutex
}

func main() {
	// Initialize the NeoPixel.
	neo.Configure(machine.PinConfig{Mode: machine.PinOutput})
	ws := ws2812.NewWS2812(neo)
	setupLED(ws)

	// Initialize the sensor settings.
	s := SensorSettings{
		interval:  1000, // 1 second
		ledDriver: &ws,
		stop:      false,
	}

	// Configure the serial port.
	machine.Serial.Configure(machine.UARTConfig{
		BaudRate: 115200,
	})

	// Start a goroutine to continuously read serial data.
	go s.ReceiveSerial()

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

		s.ReadTemperature()

		// Wait for the next interval. (minus the time spent above in ReadTemperature)
		time.Sleep(time.Duration(interval-sendingLedTime) * time.Millisecond)
	}
}

// Setup the LED. The NeoPixel is a bit particular: it requires green to be set first (i.e., on startup).
func setupLED(ws ws2812.Device) {
	// Green
	ws.WriteColors([]color.RGBA{green})
	// A slight pause
	time.Sleep(100 * time.Millisecond)
	// Off (black in this case)
	ws.WriteColors([]color.RGBA{black})
}

// ReadTemperature reads and converts the raw temperature reading.
// Also sends data on serial port.
func (s *SensorSettings) ReadTemperature() {
	// Set the LED to the sending color to indicate activity.
	s.ledDriver.WriteColors([]color.RGBA{sending})

	// Read the temperature.
	temp := float32(machine.ReadTemperature()) / 1000

	// Send the temperature data over serial.
	// Note: println in tinygo will send to serial by default.
	println(temp)

	// Wait a bit to allow the color to be seen.
	time.Sleep(sendingLedTime * time.Millisecond)

	// Set the LED to green to indicate the sensor is running.
	s.ledDriver.WriteColors([]color.RGBA{on})
}

// receiveSerial continuously reads incoming data from the serial port.
func (s *SensorSettings) ReceiveSerial() {

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
			// interval X: Set the read (and send) interval to X milliseconds, minimum 10 ms.
			// stop: Stop sending temperature data.
			// start: Start sending temperature data.
			// Note: for simplicity, for bad commands or malformed ones, we just ignore them.
			command := strings.Fields(string(buf))
			if len(command) > 0 {
				switch command[0] {
				case "interval":
					if len(command) > 1 {
						// Parse the interval value.
						interval, err := strconv.Atoi(command[1])
						if err == nil {
							s.SetInterval(interval)
						}
					}
				case "stop":
					// Stop sending temperature data.
					s.Stop()
				case "start":
					// Start sending temperature data.
					s.Start()
				}
			}
		}

		// Yield to other goroutines. (Needed in TinyGo to allow other goroutines to run.)
		runtime.Gosched()
	}
}

// "Stop" the Sensor
func (s *SensorSettings) Stop() {
	s.lock.Lock()
	s.stop = true
	s.lock.Unlock()

	// Set the LED to red to indicate the sensor is stopped.
	s.ledDriver.WriteColors([]color.RGBA{off})
}

// "Start" the Sensor
func (s *SensorSettings) Start() {
	s.lock.Lock()
	s.stop = false
	s.lock.Unlock()

	// Set the LED to green to indicate the sensor is running.
	s.ledDriver.WriteColors([]color.RGBA{on})
}

// Set the interval for reading the sensor
func (s *SensorSettings) SetInterval(interval int) {
	s.lock.Lock()
	if interval < minInterval {
		s.interval = minInterval
	} else {
		s.interval = interval
	}
	s.lock.Unlock()
}
