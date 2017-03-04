#!/bin/sh

# dependencies

apt-get update
apt-get install -y bc build-essential gcc-aarch64-linux-gnu git unzip qemu-user-static


mkdir -p build
cd build

# build kernel

git clone --depth=1 -b rpi-4.9.y https://github.com/raspberrypi/linux.git

cd linux
make ARCH=arm64 CROSS_COMPILE=aarch64-linux-gnu- bcmrpi3_defconfig
echo "CONFIG_KEYS_COMPAT=y" >> .config
make -j 3 ARCH=arm64 CROSS_COMPILE=aarch64-linux-gnu-
cd ..



# build rootfs

wget https://downloads.raspberrypi.org/raspbian_lite/images/raspbian_lite-2017-02-27/2017-02-16-raspbian-jessie-lite.zip
unzip 2017-02-16-raspbian-jessie-lite.zip

mkdir -p mnt
mount -o loop,offset=70254592 2017-02-16-raspbian-jessie-lite.img mnt
rm -rf mnt/*
multistrap -a arm64 -d $PWD/mnt -f ../multistrap.conf

cat > mnt/etc/fstab <<EOL
proc            /proc           proc    defaults          0       0
/dev/mmcblk0p1  /boot           vfat    defaults          0       2
/dev/mmcblk0p2  /               ext4    defaults,noatime  0       1
EOL

mount -o bind /dev mnt/dev/
mount -t proc proc mnt/proc
mount -t sysfs sys mnt/sys

cp /usr/bin/qemu-aarch64-static mnt/usr/bin/qemu-aarch64-static

chroot mnt /var/lib/dpkg/info/dash.preinst install
chroot mnt dpkg --configure -a

rm mnt/usr/bin/qemu-aarch64-static

sed -i 's/root:x/root:/' mnt/etc/passwd

echo raspberrypi > mnt/etc/hostname

echo 127.0.1.1 raspberrypi >> mnt/etc/hosts

cat >> mnt/etc/network/interfaces <<EOL
auto eth0
iface eth0 inet dhcp
EOL

useradd -m -p $(perl -e 'print crypt("raspberry", "password")') pi
echo "pi ALL=(ALL) NOPASSWD: ALL" > /etc/sudoers.d/010_pi-nopasswd



# install boot stuff

mount -o loop,offset=4194304,sizelimit=66060288 2017-02-16-raspbian-jessie-lite.img ../mnt/boot

sed -i 's/quiet init=\/usr\/lib\/raspi-config\/init_resize.sh//' ../mnt/boot/cmdline.txt

cd linux
cp arch/arm64/boot/Image ../mnt/boot/kernel8.img
echo "kernel=kernel8.img" >> ../mnt/boot/config.txt
cp arch/arm64/boot/dts/broadcom/bcm2710-rpi-3-b.dtb ../mnt/boot/
make ARCH=arm64 CROSS_COMPILE=aarch64-linux-gnu- INSTALL_MOD_PATH=$(dirname $PWD)/mnt modules_install
cd ..



# compress image

umount mnt/boot mnt/dev mnt/proc mnt/sys mnt
mv 2017-02-16-raspbian-jessie-lite.img pi64.img
tar -zcvf pi64.img.tar.gz pi64.img

