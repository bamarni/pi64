package main

import (
	"fmt"
	"os/exec"

	"github.com/bamarni/pi64/pkg/dialog"
	"github.com/bamarni/pi64/pkg/util"
	"github.com/bamarni/pi64/pkg/vchiq"
)

func checkCPUFreq() {
	util.AttachCommand("/usr/bin/dialog", "--infobox", "Running CPU frequency test... (this should take ~5 seconds)", "10", "80")
	if err := exec.Command("/usr/bin/sysbench", "--test=cpu", "--cpu-max-prime=10000", "--num-threads=4", "run").Run(); err != nil {
		dialog.Message("Couldn't run benchmark.")
		return
	}

	throttled, err := vchiq.GetThrottled()
	if err != nil {
		dialog.Message(err.Error())
		return
	}
	if throttled != 0 {
		var reason []string
		if throttled&(vchiq.UnderVoltage|vchiq.UnderVoltageOccured) != 0 {
			reason = append(reason, "under-voltage")
		}
		if throttled&(vchiq.FreqCap|vchiq.FreqCapOccured) != 0 {
			reason = append(reason, "arm frequency capped")
		}
		if throttled&(vchiq.Throttling|vchiq.Throttled) != 0 {
			reason = append(reason, "throttled")
		}
		dialog.Message(fmt.Sprintf("Throttling occured %v, your RPI doesn't perform well under load.\n\nThis usually happens because of a suboptimal power supply cable.", reason))
		return
	}

	dialog.Message("Congratulations! Your RPI performs well under load.")
}
