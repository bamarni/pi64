package vchiq

import (
	"errors"
	"os/exec"
	"strconv"
)

const (
	// vcgencmd get_throttled result
	UnderVoltage        int64 = 1
	FreqCap                   = 1 << 1
	Throttling                = 1 << 2
	UnderVoltageOccured       = 1 << 16
	FreqCapOccured            = 1 << 17
	Throttled                 = 1 << 18
)

func GetThrottled() (int64, error) {
	rawThrottled, err := exec.Command("vcgencmd", "get_throttled").Output()
	len := len(rawThrottled)
	if err != nil || len < 14 {
		return 0, errors.New("couldn't run vcgencmd")
	}
	rawThrottled = rawThrottled[12 : len-1]

	throttled, err := strconv.ParseInt(string(rawThrottled), 16, 32)
	if err != nil || len < 14 {
		return 0, errors.New("couldn't parse throttled output : " + string(rawThrottled))
	}
	return throttled, nil
}
