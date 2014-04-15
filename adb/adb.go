package adb

import (
	"log"
	"os"
	"os/exec"
	"strings"
	"sync"
)

func Exec(args ...string) ([]byte, error) {
	return exec.Command(os.ExpandEnv("$ANDROID_HOME/platform-tools/adb"), args...).CombinedOutput()
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
	out, err := Exec("devices")

	if err != nil {
		log.Fatal(err)
	}

	lines := strings.Split(string(out), "\n")
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
