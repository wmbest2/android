package adb

import (
	"bufio"
    "errors"
	"fmt"
	"io/ioutil"
    "net"
	"os"
	"os/exec"
	"strings"
    "strconv"
	"sync"
)

const (
    dalvikWarning = "WARNING: linker: libdvm.so has text relocations. This is wasting memory and is a security risk. Please fix."
)

type Adb struct {
    Host string
    Port int
}

func Default() *Adb {
    return &Adb{"localhost", 5037}
}

func (a *Adb) getConnection() (net.Conn, error) {
    h := fmt.Sprintf("%s:%d", a.Host, a.Port)
    return net.Dial("tcp", h)
}

func (a *Adb) readSize(reader *bufio.Reader, bcount int) (uint64, error) {
    size := make([]byte, bcount);
    reader.Read(size);
    return strconv.ParseUint(string(size), 16, 0)
}

func Devices() []byte {
    return Default().Devices();
}

func (a *Adb) Send(conn net.Conn, cmd string) (*bufio.Reader, error) {
    fmt.Fprintf(conn, "%04x%s", len(cmd), cmd)
    
    reader := bufio.NewReader(conn)
    status := make([]byte, 4);
    _, err := reader.Read(status)
    if err != nil || string(status) != `OKAY` {
        return nil, errors.New(`invalid connection`)
    }

    return reader, nil
}

func (adb *Adb) Devices() []byte {
    conn, _ := adb.getConnection()
    defer conn.Close()

    reader, _ := adb.Send(conn, "host:devices")
    size, _ := adb.readSize(reader, 4);

    lines := make([]byte, size);
    reader.Read(lines)

    return lines
}


func Exec(args ...string) chan interface{} {
	out := make(chan interface{})

	go func() {
		defer close(out)
		cmd := exec.Command(os.ExpandEnv("$ANDROID_HOME/platform-tools/adb"), args...)
		stdOut, err := cmd.StdoutPipe()
		stdErr, err := cmd.StderrPipe()

		if err != nil {
			out <- err
			return
		}

		if err != nil {
			out <- err
			return
		}

		scanner := bufio.NewScanner(stdOut)

		if err = cmd.Start(); err != nil {
			out <- err
			return
		}

		for scanner.Scan() {
            b := scanner.Bytes()
            if string(b[0:7]) == "WARNING" && string(b) == dalvikWarning {
                continue
            }
			out <- append(b, byte('\n'))
		}

		e, _ := ioutil.ReadAll(stdErr)

		if err = cmd.Wait(); err != nil {
			out <- e
			out <- err
		}
	}()

	return out
}

func ExecSync(args ...string) ([]byte, error) {
	var output []byte 
	var v interface{}
	var err error

	out := Exec(args...)
	out_ok := true

	for {
		if !out_ok {
			break
		}
		switch v, out_ok = <-out; v.(type) {
		case []byte:
			output = append(output, v.([]byte)...)
		case error:
			err = v.(error)
		}
	}
	return output, err
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

			d := &Device{Serial: device}
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
