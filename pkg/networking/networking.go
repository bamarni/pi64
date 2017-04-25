package networking

import (
	"bufio"
	"bytes"
	"io/ioutil"
	"os/exec"
	"regexp"
	"strings"
)

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

func ScanSSIDs() ([]string, error) {
	var out bytes.Buffer
	cmd := exec.Command("iwlist", "wlan0", "scan")
	cmd.Stdout = &out
	if err := cmd.Run(); err != nil {
		return nil, err
	}

	var ssids []string
	scanner := bufio.NewScanner(&out)
	for scanner.Scan() {
		line := strings.TrimLeft(scanner.Text(), " ")
		if !strings.HasPrefix(line, "ESSID") {
			continue
		}
		if splits := strings.Split(line, `"`); len(splits) == 3 {
			ssids = append(ssids, splits[1])
		}
	}
	return ssids, nil
}
