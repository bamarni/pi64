package main

import (
	"github.com/bamarni/pi64/pkg/dialog"
	"github.com/bamarni/pi64/pkg/util"
)

func configureTimezone() {
	if err := util.AttachCommand("/usr/sbin/dpkg-reconfigure", "tzdata"); err != nil {
		dialog.Message("Couldn't configure timezone.")
		return
	}
	dialog.Message("Timezone has been configured.")
}
