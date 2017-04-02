package main

import (
	"fmt"
	"os"
)

func main() {
	if os.Getpid() == 1 {
		initSetup()
		return
	}

	fmt.Println("PI64 CLI tool")
}
