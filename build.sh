#!/bin/sh

set -ex

# dependencies

apt-get update
apt-get install -y bc build-essential cmake gcc-aarch64-linux-gnu g++-aarch64-linux-gnu git unzip qemu-user-static multistrap zip wget


mkdir -p build
cd build



# build and partition image

fallocate -l 1024M pi64.img

fdisk pi64.img <<EOF
o
n


8192
137215
p
t
c
n


8192


p
w
EOF

parted_out=$(parted -s pi64.img unit b print)

boot_offset=$(echo "$parted_out" | grep -e '^ 1'| xargs echo -n | cut -d" " -f 2 | tr -d B)
boot_length=$(echo "$parted_out" | grep -e '^ 1'| xargs echo -n | cut -d" " -f 4 | tr -d B)

root_offset=$(echo "$parted_out" | grep -e '^ 2'| xargs echo -n | cut -d" " -f 2 | tr -d B)
root_length=$(echo "$parted_out" | grep -e '^ 2'| xargs echo -n | cut -d" " -f 4 | tr -d B)

boot_dev=$(losetup --show -f -o ${boot_offset} --sizelimit ${boot_length} pi64.img)
root_dev=$(losetup --show -f -o ${root_offset} --sizelimit ${root_length} pi64.img)

mkdosfs -n boot -F 32 $boot_dev
mkfs.ext4 -O ^huge_file $root_dev



# build rootfs

mkdir -p mnt
mount $root_dev mnt -t ext4

multistrap -a arm64 -d $PWD/mnt -f ../multistrap.conf

cp /usr/bin/qemu-aarch64-static mnt/usr/bin/qemu-aarch64-static

chroot mnt apt-get clean

rm -rf mnt/var/lib/apt/lists/*

cat > mnt/etc/apt/sources.list <<EOL
deb http://deb.debian.org/debian stretch main contrib non-free
deb-src http://deb.debian.org/debian stretch main contrib non-free

deb http://deb.debian.org/debian stretch-updates main contrib non-free
deb-src http://deb.debian.org/debian stretch-updates main contrib non-free

deb http://security.debian.org/ stretch/updates main contrib non-free
deb-src http://security.debian.org/ stretch/updates main contrib non-free
EOL

rm mnt/etc/apt/sources.list.d/multistrap-debian.list mnt/usr/bin/qemu-aarch64-static

cat > mnt/etc/fstab <<EOL
proc            /proc           proc    defaults          0       0
/dev/mmcblk0p1  /boot           vfat    defaults          0       2
/dev/mmcblk0p2  /               ext4    defaults,noatime  0       1
EOL

echo raspberrypi > mnt/etc/hostname

echo nameserver 8.8.8.8 > mnt/etc/resolv.conf

cat >> mnt/etc/network/interfaces <<EOL
allow-hotplug eth0
iface eth0 inet manual

allow-hotplug wlan0
iface wlan0 inet manual
    wpa-conf /etc/wpa_supplicant/wpa_supplicant.conf
EOL

cat > mnt/etc/wpa_supplicant/wpa_supplicant.conf <<EOL
ctrl_interface=DIR=/var/run/wpa_supplicant GROUP=netdev
update_config=1
EOL

chmod 600 mnt/etc/wpa_supplicant/wpa_supplicant.conf

[ ! -d ./userland ] && git clone --depth=1 https://github.com/raspberrypi/userland

mkdir -p ./userland/build && cd ./userland/build
cmake -DCMAKE_SYSTEM_NAME=Linux -DCMAKE_BUILD_TYPE=release -DARM64=ON -DCMAKE_C_COMPILER=aarch64-linux-gnu-gcc -DCMAKE_CXX_COMPILER=aarch64-linux-gnu-g++ -DCMAKE_ASM_COMPILER=aarch64-linux-gnu-gcc -DVIDEOCORE_BUILD_DIR=/opt/vc ../
make -j $(nproc)
cd ../../
mkdir -p mnt/opt && mv /opt/vc mnt/opt/
mv mnt/opt/vc/bin/* mnt/usr/bin/



# build kernel and boot stuff

mkdir -p mnt/boot
mount $boot_dev mnt/boot -t vfat

[ ! -d ./firmware ] && git clone --depth=1 https://github.com/raspberrypi/firmware

cp -r firmware/boot/* mnt/boot
echo "dwc_otg.lpm_enable=0 console=serial0,115200 console=tty1 root=/dev/mmcblk0p2 rootfstype=ext4 elevator=deadline fsck.repair=yes rootwait loglevel=3 net.ifnames=0 init=/usr/bin/pi64-config" > mnt/boot/cmdline.txt

[ ! -d ./linux ] && git clone --depth=1 -b rpi-4.11.y https://github.com/raspberrypi/linux.git

cd linux
sed -i 's/^EXTRAVERSION =.*/EXTRAVERSION = +pi64/g' Makefile
cp ../../.config ./
make ARCH=arm64 CROSS_COMPILE=aarch64-linux-gnu- olddefconfig
make -j $(nproc) ARCH=arm64 CROSS_COMPILE=aarch64-linux-gnu-
cp arch/arm64/boot/Image ../mnt/boot/kernel8.img
cp arch/arm64/boot/dts/broadcom/bcm2710-rpi-3-b.dtb ../mnt/boot/
make ARCH=arm64 CROSS_COMPILE=aarch64-linux-gnu- INSTALL_MOD_PATH=$(dirname $PWD)/mnt modules_install
cd ..

# https://github.com/RPi-Distro/repo/issues/51
mkdir -p mnt/lib/firmware/brcm
wget -P mnt/lib/firmware/brcm https://github.com/RPi-Distro/firmware-nonfree/raw/master/brcm80211/brcm/brcmfmac43430-sdio.txt


# build pi64 cli tool

if [ ! -d /usr/local/go ]; then
	wget https://storage.googleapis.com/golang/go1.8.linux-amd64.tar.gz
	echo "53ab94104ee3923e228a2cb2116e5e462ad3ebaeea06ff04463479d7f12d27ca  go1.8.linux-amd64.tar.gz" | sha256sum -c
	tar -C /usr/local -xzf go1.8.linux-amd64.tar.gz
fi

cd ..
export PATH=$PATH:/usr/local/go/bin
GOOS=linux GOARCH=arm64 go build -o ./build/mnt/usr/bin/pi64-config ./cmd/pi64-config
cd build



# compress image

umount mnt/boot mnt
losetup -d $boot_dev $root_dev
zip pi64.zip pi64.img
