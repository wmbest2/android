package adb

import (
	"strings"
	"sync"
)

const (
	dalvikWarning = "WARNING: linker: libdvm.so has text relocations. This is wasting memory and is a security risk. Please fix."
)

type Adb struct {
	Dialer
}

var (
	Default = &Adb{Dialer: Dialer{"localhost", 5037}}
)

func Devices() []byte {
	return Default.Devices()
}

func (a *Adb) Transport(conn *AdbConn) {
	conn.TransportAny()
}

func Shell(t Transporter, args ...string) chan []byte {
	out := make(chan []byte)

	go func() {
		defer close(out)
		conn, err := t.Dial()
		defer conn.Close()

		if err != nil {
			return
		}

		t.Transport(conn)
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

	conn.Write([]byte("host:devices"))
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

		conn.Write([]byte("host:track-devices"))

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

func FindDevice(serial string) Device {
	var dev Device
	devices := FindDevices(serial)
	if len(devices) > 0 {
		dev = *devices[0]
	}
	return dev
}

func FindDevices(serial ...string) []*Device {
	filter := &DeviceFilter{}
	filter.Serials = serial
	filter.MaxSdk = LATEST
	return ListDevices(filter)
}

func ListDevices(filter *DeviceFilter) []*Device {
	output := Devices()
	lines := strings.Split(string(output), "\n")

	devices := make([]*Device, 0, len(lines))

	var wg sync.WaitGroup

	for _, line := range lines {
		if strings.TrimSpace(line) != "" {
			device := strings.Split(line, "\t")[0]

			d := &Device{Dialer: Default.Dialer, Serial: device}
			devices = append(devices, d)

			wg.Add(1)
			go func() {
				defer wg.Done()
				d.Update()
			}()
		}
	}

	wg.Wait()

	result := make([]*Device, 0, len(lines))
	for _, device := range devices {
		if device.MatchFilter(filter) {
			result = append(result, device)
		}
	}

	return result
}
