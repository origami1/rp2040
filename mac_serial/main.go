package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"time"

	"go.bug.st/serial"
)

func main() {
	// Port name of my RP2040 when connected to my Mac.
	portName := "/dev/cu.usbmodem14301"

	mode := &serial.Mode{
		BaudRate: 115200,
		DataBits: 8,
		Parity:   serial.NoParity,
		StopBits: serial.OneStopBit,
	}
	port, err := serial.Open(portName, mode)
	if err != nil {
		log.Fatal(err)
	}
	defer port.Close()

	// Read from the serial port
	go func() {
		reader := bufio.NewReader(port)
		for {
			line, err := reader.ReadString('\n')
			if err != nil {
				log.Println("Error reading from serial port:", err)
				time.Sleep(time.Second)
				continue
			}
			fmt.Print(line)
			// Alternatively, data can be forwared to other processes or services, possibly
			// after some processing.
		}
	}()

	// Simulate sending commands to the microcontroller by accepting user input to send to the serial port
	for {
		reader := bufio.NewReader(os.Stdin)
		line, err := reader.ReadString('\n')
		_, err = port.Write([]byte(line))
		// Don't bother checking for number of bytes written.
		if err != nil {
			log.Println("Error writing to serial port:", err)
		}
	}

	// Runs until Ctrl+C is pressed.
}
