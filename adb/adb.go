package adb

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"time"
)

const (
	dalvikWarning = "WARNING: linker: libdvm.so has text relocations. This is wasting memory and is a security risk. Please fix."
)

type Adb struct {
	Dialer
	Method Transport
}

var (
	Default = &Adb{Dialer{"localhost", 5037}, Any}
)

func Connect(host string, port int) *Adb {
	return &Adb{Dialer{host, port}, Any}
}

func Devices() []byte {
	return Default.Devices()
}

func WaitFor(t Transporter) {
	for {
		conn, _ := t.Dial()
		err := t.Transport(conn)

		if err == nil {
			return
		}

		defer conn.Close()
	}
}

func Log(t Transporter, args ...string) chan []byte {
	out := make(chan []byte)

	go func(out chan []byte) {
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
		conn.Log(args...)

		for {
			line, _, err := conn.r.ReadLine()
			if err != nil {
				break
			}
			out <- line
		}
	}(out)
	return out
}

func Shell(t Transporter, args ...string) chan []byte {
	out := make(chan []byte)

	go func(out chan []byte) {
		defer close(out)
		conn, err := t.Dial()

		if err != nil {
			fmt.Println(err)
			return
		}

		defer conn.Close()

		err = t.Transport(conn)
		if err != nil {
			fmt.Println(err)
			fmt.Println("more than one device or emulator")
			os.Exit(2)
		}
		conn.Shell(args...)

		for {
			line, _, err := conn.r.ReadLine()
			line = bytes.Replace(line, []byte{'\r'}, []byte{}, 0)
			if err != nil {
				break
			}
			out <- line
		}
	}(out)
	return out
}

func ShellSync(t Transporter, args ...string) []byte {
	out := Shell(t, args...)
	output := make([]byte, 0)
	for line := range out {
		output = append(output, line...)
		output = append(output, '\n')
	}
	return output
}

func (a *Adb) Transport(conn *AdbConn) error {
	switch a.Method {
	case Usb:
		return conn.TransportUsb()
	case Emulator:
		return conn.TransportEmulator()
	default:
		return conn.TransportAny()
	}
}

func timeTrack(start time.Time, name string) {
	elapsed := time.Since(start)
	log.Printf("%s took %s", name, elapsed)
}

func Frame(t Transporter) []byte {
	conn, err := t.Dial()
	if err != nil {
		return []byte{}
	}
	defer conn.Close()
	err = t.Transport(conn)
	if err != nil {
		fmt.Println(err)
		fmt.Println("more than one device or emulator")
		os.Exit(2)
	}

	_, err = conn.WriteCmd("framebuffer:")

	if err != nil {
		panic(err)
	}

	version := conn.readUint32()
	fmt.Printf("Version, %d\n", version)
	depth := conn.readUint32()
	fmt.Printf("Depth, %d\n", depth)
	size := conn.readUint32()
	fmt.Println("Size, ", size)
	height := conn.readUint32()
	fmt.Println("height, ", height)
	width := conn.readUint32()
	fmt.Println("width, ", width)
	ro := conn.readUint32()
	fmt.Println("ro, ", ro)
	rl := conn.readUint32()
	fmt.Println("rl ", rl)
	bo := conn.readUint32()
	fmt.Println("bo, ", bo)
	bl := conn.readUint32()
	fmt.Println("bl ", bl)
	gro := conn.readUint32()
	fmt.Println("gro, ", gro)
	gl := conn.readUint32()
	fmt.Println("gl ", gl)
	ao := conn.readUint32()
	fmt.Println("ao, ", ao)
	al := conn.readUint32()
	fmt.Println("al ", al)

	var lines []byte
	for i := 0; i < 3; i++ {
		lines = conn.readImageBytes(size)
		conn.Write([]byte{1})
	}
	return lines
}

func (conn *AdbConn) readImageBytes(size uint32) []byte {
	defer timeTrack(time.Now(), "readImageBytes")
	lines := make([]byte, size)
	w := bytes.NewBuffer(lines)
	total, _ := w.ReadFrom(conn)
	log.Println("PUlled ", total)

	return w.Bytes()
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

		conn, err := adb.Dial()

		if err != nil {
			return
		}

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
