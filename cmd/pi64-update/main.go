package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path"
	"strings"

	"github.com/bamarni/pi64/pkg/pi64"
	"github.com/cheggaaa/pb"
	"golang.org/x/crypto/openpgp"
)

var (
	tarPath = "/root/linux.tar.gz"
	pubKey  = `
-----BEGIN PGP PUBLIC KEY BLOCK-----
Version: GnuPG v2

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
02nz
=3TLO
-----END PGP PUBLIC KEY BLOCK-----
`
)

func main() {
	os.Exit(run())
}

func run() int {
	if os.Geteuid() != 0 {
		fmt.Fprintln(os.Stderr, "pi64-update must be run as root")
		return 1
	}

	keyring, err := openpgp.ReadArmoredKeyRing(strings.NewReader(pubKey))
	if err != nil {
		fmt.Fprintln(os.Stderr, "Couldn't read keyring : "+err.Error())
		return 1
	}

	client := new(http.Client)
	client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	}

	latestReleaseResp, err := client.Get("https://github.com/bamarni/pi64-kernel/releases/latest")
	if err != nil {
		fmt.Fprintln(os.Stderr, "Couldn't request for latest release : "+err.Error())
		return 1
	}

	latestRelease, err := latestReleaseResp.Location()
	if err != nil {
		fmt.Fprintln(os.Stderr, "Couldn't parse latest release location : "+err.Error())
		return 1
	}

	metadata, err := pi64.ReadMetadata()
	if err != nil {
		fmt.Fprintln(os.Stderr, "Couldn't read pi64 metadata.")
		return 1
	}

	currentVersion := metadata.KernelVersion
	if currentVersion == "" {
		currentVersion = metadata.Version
	}

	latestVersion := path.Base(latestRelease.String())

	if latestVersion <= currentVersion {
		fmt.Fprintln(os.Stderr, "You're already using the latest version.")
		return 0
	}

	releaseEndpoint := "https://github.com/bamarni/pi64-kernel/releases/download/" + latestVersion

	fmt.Fprintf(os.Stderr, "Downloading '%s' release.\n", latestRelease)

	tarResp, err := http.Get(releaseEndpoint + "/linux.tar.gz")
	if err != nil {
		fmt.Fprintln(os.Stderr, "Couldn't get linux.tar.gz : "+err.Error())
		return 1
	}
	defer tarResp.Body.Close()

	tarFile, err := os.Create(tarPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Couldn't create %s : %s\n", tarPath, err)
		return 1
	}
	defer tarFile.Close()
	defer os.Remove(tarPath)

	sig, err := http.Get(releaseEndpoint + "/linux.tar.gz.sig")
	if err != nil {
		fmt.Fprintln(os.Stderr, "Couldn't get linux.tar.gz.sig : "+err.Error())
		return 1
	}
	defer sig.Body.Close()

	bar := pb.New64(tarResp.ContentLength).SetUnits(pb.U_BYTES)
	bar.Start()

	// Wrap the response body to :
	// - a TeeReader to check against the PGP signature
	// - a proxy reader for the progress bar display
	reader := bar.NewProxyReader(io.TeeReader(tarResp.Body, tarFile))

	_, err = openpgp.CheckDetachedSignature(keyring, reader, sig.Body)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Couldn't verify signature : "+err.Error())
		return 1
	}

	if err := exec.Command("tar", "-zxvf", tarPath, "-C", "/").Run(); err != nil {
		fmt.Fprintln(os.Stderr, "Couldn't extract "+tarPath+" : "+err.Error())
		return 1
	}

	metadata.KernelVersion = latestVersion
	if err := pi64.WriteMetadata(metadata); err != nil {
		fmt.Fprintln(os.Stderr, "Couldn't write metadata : "+err.Error())
	}

	fmt.Fprintln(os.Stderr, "Your kernel has been updated! You'll have to reboot for this to take effect.")
	return 0
}
