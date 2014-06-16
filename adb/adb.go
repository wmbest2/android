package adb

import (
	"strings"
	"sync"
)

const (
	dalvikWarning = "WARNING: linker: libdvm.so has text relocations. This is wasting memory and is a security risk. Please fix."
)

type Adb struct {
	Host string
	Port int
}

type Transporter interface {
	Transport(conn *AdbConn)
}

var (
	Default = &Adb{"localhost", 5037}
)

func Devices() []byte {
	return Default.Devices()
}

func (a *Adb) Transport(conn *AdbConn) {
	conn.TransportAny()
}

func (a *Adb) Shell(t Transporter, args ...string) chan interface{} {
	out := make(chan interface{})

	go func() {
		defer close(out)
		conn, err := Dial(a)
		if err != nil {
			out <- err
		}
		defer conn.Close()

		t.Transport(conn)
		conn.Shell(args...)

		for {
			line, _, err := conn.r.ReadLine()
			if err != nil {
				out <- err
				break
			}
			out <- line
		}
	}()
	return out
}

func (a *Adb) ShellSync(t Transporter, args ...string) ([]byte, error) {
	output := make([]byte, 0)
	out_ok := true
	var v interface{}
	out := a.Shell(t, args...)
	for {
		if !out_ok {
			break
		}
		switch v, out_ok = <-out; v.(type) {
		case []byte:
			output = append(output, v.([]byte)...)
		}
	}
	return output, nil
}

func (adb *Adb) Devices() []byte {
	conn, _ := Dial(adb)
	defer conn.Close()

	conn.Write([]byte("host:devices"))
	size, _ := conn.readSize(4)

	lines := make([]byte, size)
	conn.Read(lines)

	return lines
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

			d := &Device{Host: Default, Serial: device}
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
