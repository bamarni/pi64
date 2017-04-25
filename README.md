# pi64

pi64 is an experimental 64-bit OS for the Raspberry Pi 3. It is based on Debian Stretch and backed by a 4.9 Linux kernel.

## Installation

The latest image is always available in the [releases](https://github.com/bamarni/pi64/releases) section.

Once downloaded, you can follow the [official instructions](https://www.raspberrypi.org/documentation/installation/installing-images/README.md) for writing it to your SD card.

During first boot the installation process will continue for a few minutes, then the Raspberry Pi will reboot and you'll be ready to go.

## Getting started

The default user is `pi` and its password `raspberry`, it has passwordless root privileges escalation through `sudo`.

SSH is enabled by default.

Once logged in, you might want to run `sudo pi64-config` in order to get assisted with your setup!
