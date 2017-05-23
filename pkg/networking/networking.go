package networking

import (
	"bufio"
	"bytes"
	"io/ioutil"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
)

// Access Point
type AP struct {
	Name    string
	Quality int
}

func SetHostname(hostname string) error {
	if err := ioutil.WriteFile("/etc/hostname", []byte(hostname+"\n"), 0644); err != nil {
		return err
	}

	hosts, err := ioutil.ReadFile("/etc/hosts")
	if err != nil {
		return err
	}

	reg, _ := regexp.Compile(`(127\.0\.1\.1[\t ]+).*`)
	if reg.Match(hosts) {
		hosts = reg.ReplaceAll(hosts, []byte("${1}"+hostname))
	} else {
		hosts = append([]byte("127.0.1.1 "+hostname+"\n"), hosts...)
	}

	return ioutil.WriteFile("/etc/hosts", hosts, 0644)
}

func Ifup(iface string) error {
	return exec.Command("ifup", iface).Run()
}

func Ifdown(iface string) error {
	return exec.Command("ifdown", iface).Run()
}

// ScanAPs searches for available access points through a given wireless interface
func ScanAPs(iface string) ([]*AP, error) {
	var out bytes.Buffer
	cmd := exec.Command("iwlist", iface, "scan")
	cmd.Stdout = &out
	if err := cmd.Run(); err != nil {
		return nil, err
	}
	scanner := bufio.NewScanner(&out)

	var ap *AP
	var aps []*AP
	for scanner.Scan() {
		line := strings.TrimLeft(scanner.Text(), " ")
		if strings.HasPrefix(line, "Cell") {
			if ap != nil {
				aps = append(aps, ap)
			}
			ap = &AP{}
		} else if strings.HasPrefix(line, "ESSID") {
			if splits := strings.Split(line, `"`); len(splits) == 3 {
				ap.Name = splits[1]
			}
		} else if strings.HasPrefix(line, "Quality") {
			ap.Quality, _ = strconv.Atoi(line[8:10])
		}
	}
	aps = append(aps, ap)
	return aps, nil
}
