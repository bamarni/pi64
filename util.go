package main

import (
	"io/ioutil"
	"regexp"
)

func setHostname(hostname string) error {
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
