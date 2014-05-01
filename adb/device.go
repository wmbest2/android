package adb

import (
	"fmt"
	"strconv"
	"strings"
)

type DeviceType int
type SdkVersion int

const (
	PHONE DeviceType = iota
	TABLET_7
	TABLET_10
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
	LATEST SdkVersion = iota - 1
)

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
}

type AdbRunner interface {
	Exec(args ...string) chan interface{}
	ExecSync(args ...string) ([]byte, error)
}

type Device struct {
	Serial       string     `json:"serial"`
	Manufacturer string     `json:"manufacturer"`
	Model        string     `json:"model"`
	Sdk          SdkVersion `json:"sdk"`
	Version      string     `json:"version"`
    Density      int64      `json:"density"`
}

type DeviceFilter struct {
	Type    DeviceType
	Serials []string
    Density int64
	MinSdk  SdkVersion
	MaxSdk  SdkVersion
}

func (s SdkVersion) String() string {
	return sdkMap[s]
}

// filter -f "serials=[...];type=tablet;count=5;version >= 4.1.1;"

/*func GetFilter(arg string) {*/

/*}*/
func (d *Device) Exec(args ...string) chan interface{} {
	args = append([]string{"-s", d.Serial}, args...)
	return Exec(args...)
}

func (d *Device) ExecSync(args ...string) ([]byte, error) {
	args = append([]string{"-s", d.Serial}, args...)
	return ExecSync(args...)
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

	if d.Sdk < filter.MinSdk || (filter.MaxSdk != 0 && d.Sdk > filter.MaxSdk) {
		return false
	} else if !stringInSlice(d.Serial, filter.Serials) {
		return false
    } else if filter.Density != 0 && filter.Density != d.Density{
        return false
	}
	return true
}

func (d *Device) GetProp(prop string) chan string {
	out := make(chan string)
	go func() {
		p, err := d.ExecSync("shell", "getprop", prop)
		if err == nil {
			out <- strings.TrimSpace(string(p))
		} else {
			out <- ""
		}
	}()

	return out
}

func (d *Device) SetScreenOn(on bool) {
	screen, err := d.ExecSync("shell", "dumpsys", "input_method", "|", "grep", "mScreenOn=false")
	if err != nil {
		return
	}

    if screen != nil && on || screen == nil && !on{
        d.SendKey(26)
    }
}

func (d *Device) SendKey(aKey int) {
    d.ExecSync("shell", "input", "keyevent", fmt.Sprintf("%d", aKey))
}

func (d *Device) Unlock() {
    d.SendKey(82)
}

func (d *Device) Update() {

	d.Manufacturer = <-d.GetProp("ro.product.manufacturer")
	d.Model = <-d.GetProp("ro.product.model")
	d.Version = <-d.GetProp("ro.build.version.release")
	sdk := <-d.GetProp("ro.build.version.sdk")
	den := <-d.GetProp("ro.sf.lcd_density")

	sdk_int, _ := strconv.ParseInt(sdk, 10, 0)
	d.Sdk = SdkVersion(sdk_int)

	d.Density, _ = strconv.ParseInt(den, 10, 0)
}

func (d *Device) String() string {
	return fmt.Sprintf("%s\t%s %s\t[%s (%s)]", d.Serial, d.Manufacturer, d.Model, d.Version, sdkMap[d.Sdk])
}
