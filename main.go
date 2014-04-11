package main

import (
    "fmt"
    "flag"
)

func runOnAll(params []string) []byte {
    devices := AdbDevices(nil)
    var out string
    for _,d := range devices {
        fmt.Printf("%s\n", d)
        v,_ := d.AdbExec(params...)
        out += string(v) + "\n"
    }
    return []byte(out)
}

func flagFromBool(f bool, s string) *string {
    if (!f) {
        return nil
    }
    result := fmt.Sprintf("-%s", s)
    return &result
}

func main() {
    s := flag.String("s", "", "directs command to the device or emulator with the given\nserial number or qualifier. Overrides ANDROID_SERIAL\n environment variable.")
    p := flag.String("p", "", "directs command to the device or emulator with the given\nserial number or qualifier. Overrides ANDROID_SERIAL\n environment variable.")
    a := flag.Bool("a", false, "directs adb to listen on all interfaces for a connection")
    d := flag.Bool("d", false, "directs command to the only connected USB device\nreturns an error if more than one USB device is present.")
    e := flag.Bool("e", false, "directs command to the device or emulator with the given\nserial number or qualifier. Overrides ANDROID_SERIAL\n environment variable.")
    H := flag.String("H", "", "directs command to the device or emulator with the given\nserial number or qualifier. Overrides ANDROID_SERIAL\n environment variable.")
    P := flag.String("P", "", "directs command to the device or emulator with the given\nserial number or qualifier. Overrides ANDROID_SERIAL\n environment variable.")

    flag.Parse()

    aFlag := flagFromBool(*a, "a")
    dFlag := flagFromBool(*d, "d")
    eFlag := flagFromBool(*e, "e")

    allParams := []*string{aFlag,dFlag,eFlag,p,H,P}
    params := make([]string, 0, 7)
    for _, param := range allParams {
        if (param != nil && *param != "") {
            params = append(params, []string{*param}...)
        }
    }

    l := len(params) + len(flag.Args())
    args := make([]string, l, l)
    args = append(args, params...)
    args = append(args, flag.Args()...)

    var out []byte
    if (*s != "") {
        fmt.Printf("Serial: %s\n", *s)
        d := FindDevice(*s)
        out,_ = d.AdbExec(flag.Args()...)
    } else {

        if (flag.Arg(0) == "install") {
            out = runOnAll(args)
        } else {
            out,_ = AdbExec(flag.Args()...)
        }
    }
    fmt.Sprint(string(out))
}
