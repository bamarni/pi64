package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/bamarni/pi64/pkg/diskutil"
)

const (
	Lite    string = "lite"
	Desktop string = "desktop"
)

var (
	buildDir string
	image    *diskutil.Image
	bootPart *diskutil.Partition
	rootPart *diskutil.Partition
	rootDir  string
	bootDir  string
	version  string
	debug    bool
)

func main() {
	flag.StringVar(&buildDir, "build-dir", "", "Build directory")
	flag.StringVar(&version, "version", Lite, "pi64 version ('lite' or 'desktop')")
	flag.BoolVar(&debug, "debug", false, "Create a debug image")
	flag.Parse()

	if version != Lite && version != Desktop {
		fmt.Fprintln(os.Stderr, "Unsupported version "+version)
		os.Exit(1)
	}

	fmt.Fprintln(os.Stderr, "Creating build directory...")
	if err := makeBuildDir(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	fmt.Fprintln(os.Stderr, "Creating image...")
	checkError(createImage())

	fmt.Fprintln(os.Stderr, "Installing Debian...")
	checkError(installDebian())

	fmt.Fprintln(os.Stderr, "Finishing installation...")
	checkError(finishInstall())
}

func runCommand(cmd string, args ...string) error {
	out, err := exec.Command(cmd, args...).CombinedOutput()
	if err != nil {
		fmt.Fprintln(os.Stderr, string(out))
	}
	return err
}

func checkError(err error) {
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func makeBuildDir() error {
	var err error
	if buildDir, err = filepath.Abs(buildDir); err != nil {
		return err
	}
	return os.MkdirAll(buildDir, 0755)
}
