package main

import (
	"os"
	"fmt"
)

func main() {
	out,err := AdbExec(os.Args[1:]...)

	if (err != nil) {
		fmt.Printf("%s", err)
		return
	}

	for _, device := range AdbDevices() {
		fmt.Printf("D: %s\n", device)
	}

	fmt.Print(string(out))
}
