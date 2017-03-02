# dependencies

apt-get install -y bc build-essential gcc-aarch64-linux-gnu git unzip qemu-user-static



# build kernel

git clone --depth=1 -b rpi-4.8.y https://github.com/raspberrypi/linux.git
cd linux
make ARCH=arm64 CROSS_COMPILE=aarch64-linux-gnu- bcmrpi3_defconfig
make -j 3 ARCH=arm64 CROSS_COMPILE=aarch64-linux-gnu-
cd ..

# build rootfs

wget https://downloads.raspberrypi.org/raspbian_lite/images/raspbian_lite-2017-02-27/2017-02-16-raspbian-jessie-lite.zip
unzip 2017-02-16-raspbian-jessie-lite.zip

mount -o loop,offset=70254592 2017-02-16-raspbian-jessie-lite.img /mnt
rm -rf /mnt/*
multistrap -a arm64 -d /mnt -f multistrap.conf

cp fstab /mnt/etc/
cp /usr/bin/qemu-aarch64-static /mnt/usr/bin/qemu-aarch64-static

chroot /mnt

mount -t proc none /proc
dpkg --configure -a
umount /proc



# install boot stuff

mount -o loop,offset=4194304,sizelimit=66060288 2017-02-16-raspbian-jessie-lite.img /mnt/boot

cd linux
cp arch/arm64/boot/Image /mnt/boot/kernel8.img
cp arch/arm64/boot/dts/broadcom/bcm2710-rpi-3-b.dtb /mnt/boot/
echo "kernel=kernel8.img" >> /mnt/boot/config.txt
make ARCH=arm64 CROSS_COMPILE=aarch64-linux-gnu- INSTALL_MOD_PATH=/mnt modules_install

