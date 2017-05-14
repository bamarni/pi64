package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"syscall"
)

func finishInstall() error {
	script := exec.Command("bash", "-s", version, rootPart.Path())

	script.Stdin = strings.NewReader(`
set -ex

version=${1:-lite}
root_devmap=$2

cd build

rm -rf root-$version/var/lib/apt/lists/*

# install videocore

mkdir -p root-$version/opt
cp -R /opt/vc root-$version/opt/
mv root-$version/opt/vc/bin/* root-$version/usr/bin/

# install kernel and boot stuff

cp -r firmware/boot/* boot-$version

cd linux
cp arch/arm64/boot/Image ../boot-$version/kernel8.img
cp arch/arm64/boot/dts/broadcom/bcm2710-rpi-3-b.dtb ../boot-$version/
make ARCH=arm64 CROSS_COMPILE=aarch64-linux-gnu- INSTALL_MOD_PATH=$(dirname $PWD)/root-$version modules_install
cd ..

# https://github.com/RPi-Distro/repo/issues/51
mkdir -p root-$version/lib/firmware/brcm
wget -P root-$version/lib/firmware/brcm https://github.com/RPi-Distro/firmware-nonfree/raw/master/brcm80211/brcm/brcmfmac43430-sdio.txt

# build pi64 cli tool
GOOS=linux GOARCH=arm64 go build -o ./root-$version/usr/bin/pi64-config github.com/bamarni/pi64/cmd/pi64-config

# shrink and compress image

umount --lazy boot-$version root-$version

min_root_size=$(resize2fs -P $root_devmap | sed 's/Estimated minimum size of the filesystem: //')
e2fsck -fy $root_devmap
resize2fs $root_devmap $min_root_size
last_sector=$(echo "$min_root_size * 8 + 137215" | bc)
sync

kpartx -d ./pi64-$version.img

fdisk ./pi64-$version.img <<EOF
d
2
n
p
2
137216
$last_sector
w
EOF

truncate --size=$(echo "($last_sector + 1) * 512" |bc) pi64-$version.img
`)
	if out, err := script.CombinedOutput(); err != nil {
		fmt.Println(string(out))
		return err
	}

	syscall.Sync()

	if err := exec.Command("zip", buildDir+"/pi64-"+version+".zip", image.Path()).Run(); err != nil {
		return err
	}
	return os.Remove(image.Path())

	return nil
}
