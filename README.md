# pi64

pi64 is an experimental 64-bit OS for the Raspberry Pi 3. It is based on Debian Stretch and backed by a 4.11 Linux kernel.

## Releases

The latest images are always available in the [releases](https://github.com/bamarni/pi64/releases) section.

There are 2 versions : `lite` and `desktop`. The desktop version is based on [LXDE](http://lxde.org/).

## Installation

Once downloaded, you can follow [these instructions](https://www.raspberrypi.org/documentation/installation/installing-images/README.md) for writing the image to your SD card.

During first boot the installation process will continue for a few minutes, then the Raspberry Pi will reboot and you'll be ready to go.

## Getting started

The default user is `pi` and its password `raspberry`, it has passwordless root privileges escalation through `sudo`.

Once logged in, you might want to run `sudo pi64-config` in order to get assisted with your setup!

On the lite version, SSH is enabled by default.

## FAQ

### Can I still run 32-bit programs with pi64?

You should be able to run 32-bit programs out of the box as long as they're statically linked. You can check this with the `file` command : 

    $ file ./my-executable
    ./my-executable: ELF 32-bit LSB executable, ARM, EABI5 version 1 (SYSV), statically linked, not stripped

If your 32-bit program is shown as dynamically linked instead, you should still be able to run it by enabling [multiarch](https://wiki.debian.org/Multiarch/HOWTO) and installing program's required libraries :

    sudo dpkg --add-architecture armhf
    sudo apt-get update
    sudo apt-get install libc6:armhf

Here we're only installing the GNU C Library, but your program might need additional libraries.

### How can I remove SSH?

On the lite version and for convenience, SSH is installed and enabled by default. This allows you to plug your Raspberry Pi to your home router and get started without the need
of an extra monitor / keyboard. If you want to remove it, just run :

    sudo apt-get autoremove --purge -y ssh avahi-daemon

### Is there a way to run custom post-installation steps?

You can just drop a file called `setup` on the boot partition. When the installer notices that file at `/boot/setup`, it will automatically execute it using bash when installation finishes.

This can be useful if you want to distribute your own image based on pi64.
