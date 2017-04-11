package main

import (
	"io/ioutil"
	"regexp"
)

func setHostname(hostname string) error {
	if err := ioutil.WriteFile("/etc/hostname", []byte(hostname), 0644); err != nil {
		return err
	}
	hosts, err := ioutil.ReadFile("/etc/hosts")
	if err != nil {
		return err
	}
	reg, _ := regexp.Compile(`(127\.0\.1\.1[\t ]+).*`)
	return ioutil.WriteFile("/etc/hosts", reg.ReplaceAll(hosts, []byte("${1}"+hostname)), 0644)
}
