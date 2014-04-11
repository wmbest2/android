package main

import (
    "os"
    "fmt"
)

func main() {
    out,_ := AdbExec(os.Args[1:]...)

    fmt.Print(string(out))
}
