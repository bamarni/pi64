package main

import (
	"os"
)

func main() {
	if os.Getpid() == 1 {
		initSetup()
		return
	}
}
