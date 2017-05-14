package util

import (
	"os"
	"os/exec"
	"syscall"
)

func AttachCommand(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func Chroot(path string) (func() error, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	root, err := os.Open("/")
	if err != nil {
		return nil, err
	}

	if err := syscall.Chroot(path); err != nil {
		root.Close()
		return nil, err
	}

	return func() error {
		defer root.Close()
		if err := root.Chdir(); err != nil {
			return err
		}
		if err := syscall.Chroot("."); err != nil {
			return err
		}
		return os.Chdir(cwd)
	}, nil
}

func Logo() string {
	return `
                        iiii               66666666          444444444
                       i::::i             6::::::6          4::::::::4
                        iiii             6::::::6          4:::::::::4
                                        6::::::6          4::::44::::4
 ppppp   ppppppppp    iiiiiii          6::::::6          4::::4 4::::4
 p::::ppp:::::::::p   i:::::i         6::::::6          4::::4  4::::4
 p:::::::::::::::::p   i::::i        6::::::6          4::::4   4::::4
 pp::::::ppppp::::::p  i::::i       6::::::::66666    4::::444444::::444
  p:::::p     p:::::p  i::::i      6::::::::::::::66  4::::::::::::::::4
  p:::::p     p:::::p  i::::i      6::::::66666:::::6 4444444444:::::444
  p:::::p     p:::::p  i::::i      6:::::6     6:::::6          4::::4
  p:::::p    p::::::p  i::::i      6:::::6     6:::::6          4::::4
  p:::::ppppp:::::::p i::::::i     6::::::66666::::::6          4::::4
  p::::::::::::::::p  i::::::i      66:::::::::::::66         44::::::44
  p::::::::::::::pp   i::::::i        66:::::::::66           4::::::::4
  p::::::pppppppp     iiiiiiii          666666666             4444444444
  p:::::p
  p:::::p
 p:::::::p
 p:::::::p
 p:::::::p
 ppppppppp
`
}
