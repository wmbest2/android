package main

import (
	"fmt"
	"strings"
)

type Device struct {
	Serial string
	Manufacturer string
	Model string
	Version string
}

func (d *Device) AdbExec(args ...string) ([]byte, error) {
	args = append([]string{"-s", d.Serial}, args...)
	return AdbExec(args...)
}

func (d *Device) GetProp(prop string) string {
	p, err := d.AdbExec("shell", "getprop", prop)
	if (err == nil) {
		return strings.TrimSpace(string(p));
	}
	return ""
}

func (d *Device) Update() {
	d.Manufacturer = d.GetProp("ro.product.manufacturer")
	d.Model = d.GetProp("ro.product.model")
	d.Version = d.GetProp("ro.build.version.release")
}

func (d *Device) String() string {
	return fmt.Sprintf("%s %s [%s]: %s", d.Manufacturer, d.Model, d.Version, d.Serial)
}
