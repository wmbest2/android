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

	fmt.Print(string(out))
}
