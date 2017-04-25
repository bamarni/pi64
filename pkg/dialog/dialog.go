package dialog

import (
	"bytes"
	"github.com/bamarni/pi64/pkg/util"
	"os"
	"os/exec"
)

func Info(msg string) {
	util.AttachCommand("dialog", "--infobox", msg, "10", "80")
}

func Message(msg string) {
	util.AttachCommand("dialog", "--msgbox", msg, "10", "80")
}

func YesNo(msg string) bool {
	err := util.AttachCommand("dialog", "--yesno", msg, "10", "80")
	return err == nil
}

func Prompt(kind, msg string, args ...string) string {
	var res bytes.Buffer
	args = append([]string{"--" + kind, msg, "0", "80"}, args...)

	cmd := exec.Command("dialog", args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = &res

	if err := cmd.Run(); err != nil {
		return ""
	}
	return res.String()
}
