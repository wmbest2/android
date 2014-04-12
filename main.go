package main

import (
    "fmt"
    "flag"
    "sync"
)

func runOnDevice(wg *sync.WaitGroup, d Device, params []string) {
    defer wg.Done()
    fmt.Printf("%s\n", &d)
    v,_ := d.AdbExec(params...)
    fmt.Printf("%s\n", string(v))
}

func runOnAll(params []string) []byte {
    var wg sync.WaitGroup
    devices := AdbDevices(nil)
    for _,d := range devices {
        wg.Add(1)
        go runOnDevice(&wg, d, params)
    }
    wg.Wait()
    return []byte("")
}

func flagFromBool(f bool, s string) *string {
    result := fmt.Sprintf("-%s", s)
    if (!f) {
        result = ""
    }
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
        if (*param != "") {
            params = append(params, []string{*param}...)
        }
    }

    l := len(params) + len(flag.Args())
    args := make([]string, 0, l)
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
        } else if (flag.Arg(0) == "uninstall") {
            out = runOnAll(args)
        } else {
            out,_ = AdbExec(flag.Args()...)
        }
    }
    fmt.Sprint(string(out))
}
