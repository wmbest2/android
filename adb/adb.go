package adb

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"strings"
	"sync"
)

func Exec(args ...string) chan interface{} {
	out := make(chan interface{})

	go func() {
		defer close(out)
		cmd := exec.Command(os.ExpandEnv("$ANDROID_HOME/platform-tools/adb"), args...)
		stdOut, err := cmd.StdoutPipe()

		if err != nil {
			out <- err
			return
		}

		stdErr, err := cmd.StderrPipe()

		if err != nil {
			out <- err
			return
		}

		scanner := bufio.NewScanner(stdOut)
		if err = cmd.Start(); err != nil {
			e, _ := ioutil.ReadAll(stdErr)
			out <- e
			out <- err
			return
		}

		for scanner.Scan() {
			t := fmt.Sprintln(scanner.Text())
			out <- t
		}
		e, _ := ioutil.ReadAll(stdErr)

		if err = cmd.Wait(); err != nil {
			out <- string(e)
			out <- err
		}
	}()

	return out
}

func ExecSync(args ...string) ([]byte, error) {
	var output string
	var v interface{}
	var err error

	out := Exec(args...)
	out_ok := true

	for {
		if !out_ok {
			break
		}
		switch v, out_ok = <-out; v.(type) {
		case error:
			err = v.(error)
		case string:
			output = output + v.(string)
		}
	}
	return []byte(output), err
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
	output, err := ExecSync("devices")

	if err != nil {
		fmt.Println(string(output))
		log.Fatal(err)
	}

	lines := strings.Split(string(output), "\n")
	lines = lines[1:]

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
