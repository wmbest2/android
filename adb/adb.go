package adb

import (
	"fmt"
	"os"
)

const (
	dalvikWarning = "WARNING: linker: libdvm.so has text relocations. This is wasting memory and is a security risk. Please fix."
)

type Adb struct {
	Dialer
}

var (
	Default = &Adb{Dialer{"localhost", 5037}}
)

func Devices() []byte {
	return Default.Devices()
}

func (a *Adb) Transport(conn *AdbConn) error {
	return conn.TransportAny()
}

func Shell(t Transporter, args ...string) chan []byte {
	out := make(chan []byte)

	go func() {
		defer close(out)
		conn, err := t.Dial()

		if err != nil {
			fmt.Println(err)
			return
		}

		defer conn.Close()

		err = t.Transport(conn)
		if err != nil {
			fmt.Println("more than one device or emulator")
			os.Exit(2)
		}
		conn.Shell(args...)

		for {
			line, _, err := conn.r.ReadLine()
			if err != nil {
				break
			}
			out <- line
		}
	}()
	return out
}

func ShellSync(t Transporter, args ...string) []byte {
	output := make([]byte, 0)
	out := Shell(t, args...)
	for line := range out {
		output = append(output, line...)
	}
	return output
}

func (adb *Adb) Devices() []byte {
	conn, err := adb.Dial()
	if err != nil {
		return []byte{}
	}
	defer conn.Close()

	conn.WriteCmd("host:devices")
	size, _ := conn.readSize(4)

	lines := make([]byte, size)
	conn.Read(lines)

	return lines
}

func (adb *Adb) TrackDevices() chan []byte {
	out := make(chan []byte)
	go func() {
		defer close(out)

		conn, _ := adb.Dial()
		defer conn.Close()

		conn.WriteCmd("host:track-devices")

		for {
			size, err := conn.readSize(4)
			if err != nil {
				break
			}

			lines := make([]byte, size)
			_, err = conn.Read(lines)
			if err != nil {
				break
			}
			out <- lines
		}
	}()
	return out
}

func (adb *Adb) FindDevice(serial string) Device {
	var dev Device
	devices := adb.FindDevices(serial)
	if len(devices) > 0 {
		dev = *devices[0]
	}
	return dev
}

func (adb *Adb) FindDevices(serial ...string) []*Device {
	filter := &DeviceFilter{}
	filter.Serials = serial
	filter.MaxSdk = LATEST
	return adb.ListDevices(filter)
}

func ListDevices(filter *DeviceFilter) []*Device {
	return Default.ListDevices(filter)
}

func (adb *Adb) ListDevices(filter *DeviceFilter) []*Device {
	output := adb.Devices()
	return adb.ParseDevices(filter, output)
}
