package main 

import (
    "log"
	"os"
    "os/exec"
    "strings"
)

func AdbExec(args ...string) ([]byte, error) {
	return exec.Command(os.ExpandEnv("$ANDROID_HOME/platform-tools/adb"), args...).Output()
}

func AdbDevices() []Device {
	out, err := AdbExec("devices")
	
	if err != nil {
			log.Fatal(err)
	}

    lines := strings.Split(string(out), "\n")
    lines = lines[1:]

	devices := make([]Device, 0, len(lines))

    for _, line := range lines {
        if strings.TrimSpace(line) != "" {
            device := strings.Split(line, "\t")[0]
            d := &Device{Serial: device}
			d.Update();

			devices = append(devices, *d)
        }
    }

	return devices
}
