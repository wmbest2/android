package main 

import (
    "fmt"
    "log"
    "os"
    "os/exec"
    "strings"
)

func AdbExec(args ...string) ([]byte, error) {
    fmt.Printf("Args: %s\n", args)
    return exec.Command(os.ExpandEnv("$ANDROID_HOME/platform-tools/adb"), args...).CombinedOutput()
}

func FindDevice(serial string) Device {
    var dev Device
    devices := FindDevices(serial)
    if (len(devices) > 0) {
        dev = devices[0]
    }
    return dev 
}

func FindDevices(serial ...string) []Device {
    filter := &DeviceFilter{}
    filter.Serials = serial
    filter.MaxSdk = LATEST
    return AdbDevices(filter)
}

func AdbDevices(filter *DeviceFilter) []Device {
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
            if (d.MatchFilter(filter)) {
                d.Update();
                devices = append(devices, *d)
            }
        }
    }

    return devices
}
