package adb

import (
	"bytes"
	"fmt"
	"math"
	"regexp"
	"strconv"
	"strings"
	"sync"
)

type DensityBucket int
type DeviceType int
type SdkVersion int

const (
	PHONE DeviceType = iota
	TABLET_7
	TABLET_10
)

const (
	LDPI    DensityBucket = 120
	MDPI                  = 160
	HDPI                  = 240
	XHDPI                 = 320
	XXHDPI                = 480
	XXXHDPI               = 640
)

const (
	ECLAIR SdkVersion = iota + 7
	FROYO
	GINGERBREAD
	GINGERBREAD_MR1
	HONEYCOMB
	HONEYCOMB_MR1
	HONEYCOMB_MR2
	ICE_CREAM_SANDWICH
	ICE_CREAM_SANDWICH_MR1
	JELLY_BEAN
	JELLY_BEAN_MR1
	JELLY_BEAN_MR2
	KITKAT
	WEAR
	LOLLIPOP
	LATEST = LOLLIPOP
)

var typeMap = map[DeviceType]string{
	PHONE:     `Phone`,
	TABLET_7:  `7in Tablet`,
	TABLET_10: `10in Tablet`,
}

var sdkMap = map[SdkVersion]string{
	ECLAIR:                 `ECLAIR`,
	FROYO:                  `FROYO`,
	GINGERBREAD:            `GINGERBREAD`,
	GINGERBREAD_MR1:        `GINGERBREAD_MR1`,
	HONEYCOMB:              `HONEYCOMB`,
	HONEYCOMB_MR1:          `HONEYCOMB_MR1`,
	HONEYCOMB_MR2:          `HONEYCOMB_MR2`,
	ICE_CREAM_SANDWICH:     `ICE_CREAM_SANDWICH`,
	ICE_CREAM_SANDWICH_MR1: `ICE_CREAM_SANDWICH_MR1`,
	JELLY_BEAN:             `JELLY_BEAN`,
	JELLY_BEAN_MR1:         `JELLY_BEAN_MR1`,
	JELLY_BEAN_MR2:         `JELLY_BEAN_MR2`,
	KITKAT:                 `KITKAT`,
	WEAR:                   `WEAR v1`,
	LOLLIPOP:               `LOLLIPOP`,
}

type Device struct {
	Dialer       `json:"-"`
	Serial       string            `json:"serial"`
	Manufacturer string            `json:"manufacturer"`
	Model        string            `json:"model"`
	Sdk          SdkVersion        `json:"sdk"`
	Version      string            `json:"version"`
	Density      DensityBucket     `json:"density"`
	Height       int64             `json:"height"`
	Width        int64             `json:"width"`
	Properties   map[string]string `json:_`
}

type DeviceFilter struct {
	Type    DeviceType
	Serials []string
	Density DensityBucket
	MinSdk  SdkVersion
	MaxSdk  SdkVersion
}

var (
	AllDevices = &DeviceFilter{MaxSdk: LATEST}
)

func (s SdkVersion) String() string {
	return sdkMap[s]
}

// filter -f "serials=[...];type=tablet;count=5;version >= 4.1.1;"

/*func GetFilter(arg string) {*/

/*}*/

type DeviceWatcher []chan []*Device

func (d *Device) Type() DeviceType {
	sw := math.Min(float64(d.Height), float64(d.Width))
	dip := float64(LDPI) / float64(d.Density) * sw
	if dip >= 720 {
		return TABLET_10
	} else if dip >= 600 {
		return TABLET_7
	}
	return PHONE
}

func (d *Device) Transport(conn *AdbConn) error {
	return conn.TransportSerial(d.Serial)
}

func (adb *Adb) ParseDevices(filter *DeviceFilter, input []byte) []*Device {
	lines := strings.Split(string(input), "\n")

	devices := make([]*Device, 0, len(lines))

	var wg sync.WaitGroup

	for _, line := range lines {
		if strings.Contains(line, "device") && strings.TrimSpace(line) != "" {
			device := strings.Split(line, "\t")[0]

			d := &Device{Dialer: adb.Dialer, Serial: device}
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

func stringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return len(list) == 0
}

func (d *Device) MatchFilter(filter *DeviceFilter) bool {
	if filter == nil {
		return true
	}

	if d.Sdk < filter.MinSdk {
		return false
	} else if filter.MaxSdk != 0 && d.Sdk > filter.MaxSdk {
		return false
	} else if !stringInSlice(d.Serial, filter.Serials) {
		return false
	} else if filter.Density != 0 && filter.Density != d.Density {
		return false
	}
	return true
}

func (d *Device) RefreshProps() {
	d.Properties = make(map[string]string)

	proprx, err := regexp.Compile("\\[(.*)\\]: \\[(.*)\\]")

	if err != nil {
		panic(err)
	}

	out := Shell(d, "getprop")
	for line := range out {
		if line != nil {
			matches := proprx.FindSubmatch(line)
			if len(matches) > 2 {
				d.Properties[string(matches[1])] = string(matches[2])
			}
		}
	}
}

func (d *Device) GetProp(prop string) string {
	return d.Properties[prop]
}

func (d *Device) HasPackage(pack string) bool {
	return d.findValue(pack, "pm", "list", "packages", "-3")
}

func (d *Device) SetScreenOn(on bool) {
	current := d.findValue("mScreenOn=false", "dumpsys", "input_method")
	if current && on || !current && !on {
		d.SendKey(26)
	}
}

func (d *Device) findValue(val string, args ...string) bool {
	out := Shell(d, args...)
	current := false
	for line := range out {
		if line != nil {
			current = bytes.Contains(line, []byte(val))
			if current {
				break
			}
		}
	}
	return current
}

func (d *Device) SendKey(aKey int) {
	ShellSync(d, "input", "keyevent", fmt.Sprintf("%d", aKey))
}

func (d *Device) Unlock() {
	current := d.findValue("mLockScreenShown true", "dumpsys", "activity")
	if current {
		d.SendKey(82)
	}
}

func (d *Device) Update() {

	WaitFor(d)
	d.RefreshProps()

	out := []string{
		d.GetProp("ro.product.manufacturer"),
		d.GetProp("ro.product.model"),
		d.GetProp("ro.build.version.release"),
		d.GetProp("ro.build.version.sdk"),
		d.GetProp("ro.sf.lcd_density"),
	}

	d.Manufacturer = out[0]
	d.Model = out[1]
	d.Version = out[2]

	// Parse Version Code
	sdk_int, _ := strconv.ParseInt(out[3], 10, 0)
	d.Sdk = SdkVersion(sdk_int)

	// Parse DensityBucket
	density, _ := strconv.ParseInt(out[4], 10, 0)
	d.Density = DensityBucket(density)
}

func (d *Device) String() string {
	return fmt.Sprintf("%s\t%s %s\t[%s (%s) %s ]", d.Serial, d.Manufacturer, d.Model, d.Version, sdkMap[d.Sdk], typeMap[d.Type()])
}
