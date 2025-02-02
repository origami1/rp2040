# RP2040 Microcontroller Onboard Temperature Sensor
Trivial example of reading the onboard temperature sensor on an RP2040, as well as a utility to read the data over serial. Uses Go (and TinyGo).

## RP2040
Build (and deploy) using TinyGo: tinygo flash -target pico

## MacOS
Build using Go: go build -o mac_serial .

The RP2040 microcontroller supports a serial connection over USB. The serial utility allows sending commands to the RP2040 to start/stop the reading/sending of temperature data as well as to set the reading/sending interval (in seconds, defautls to 1).

