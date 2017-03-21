#!/bin/sh

set -ex

# dependencies

apt-get update
apt-get install -y bc build-essential gcc-aarch64-linux-gnu git unzip qemu-user-static multistrap zip


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

rm mnt/usr/bin/qemu-aarch64-static

cat > mnt/etc/fstab <<EOL
proc            /proc           proc    defaults          0       0
/dev/mmcblk0p1  /boot           vfat    defaults          0       2
/dev/mmcblk0p2  /               ext4    defaults,noatime  0       1
EOL

echo exit 101 > mnt/usr/sbin/policy-rc.d

chmod +x mnt/usr/sbin/policy-rc.d

echo raspberrypi > mnt/etc/hostname

echo 127.0.1.1 raspberrypi >> mnt/etc/hosts

echo nameserver 8.8.8.8 > mnt/etc/resolv.conf

cat >> mnt/etc/network/interfaces <<EOL
auto eth0
iface eth0 inet dhcp
EOL

cat > mnt/root/init_setup.sh <<EOL
#!/bin/sh

export DEBIAN_FRONTEND=noninteractive DEBCONF_NONINTERACTIVE_SEEN=true
export LC_ALL=C LANGUAGE=C LANG=C
export PATH=/usr/sbin:/usr/bin:/sbin:/bin

mount -t proc proc /proc
mount -t sysfs sys /sys
mount /boot
mount -o remount,rw /

parted /dev/mmcblk0 u s resizepart 2 \$(expr \$(cat /sys/block/mmcblk0/size) - 1)
resize2fs /dev/mmcblk0p2

/var/lib/dpkg/info/dash.preinst install
dpkg --configure -a

rm /usr/sbin/policy-rc.d

useradd -s /bin/bash --create-home -p $(perl -e 'print crypt("raspberry", "password")') pi
echo "pi ALL=(ALL) NOPASSWD: ALL" > /etc/sudoers.d/010_pi-nopasswd

sed -i 's| init=/root/init_setup.sh||' /boot/cmdline.txt
sync

echo 1 > /proc/sys/kernel/sysrq
rm /root/init_setup.sh
echo b > /proc/sysrq-trigger
EOL

chmod +x mnt/root/init_setup.sh



# build kernel and boot stuff

mkdir -p mnt/boot
mount $boot_dev mnt/boot -t vfat

git clone --depth=1 https://github.com/raspberrypi/firmware

cp -r firmware/boot/* mnt/boot
echo "kernel=kernel8.img" >> mnt/boot/config.txt
echo "dwc_otg.lpm_enable=0 console=serial0,115200 console=tty1 root=/dev/mmcblk0p2 rootfstype=ext4 elevator=deadline fsck.repair=yes rootwait net.ifnames=0 init=/root/init_setup.sh" > mnt/boot/cmdline.txt

git clone --depth=1 -b rpi-4.9.y https://github.com/raspberrypi/linux.git

cd linux
make ARCH=arm64 CROSS_COMPILE=aarch64-linux-gnu- bcmrpi3_defconfig
echo "CONFIG_KEYS_COMPAT=y" >> .config
make -j 3 ARCH=arm64 CROSS_COMPILE=aarch64-linux-gnu-
cp arch/arm64/boot/Image ../mnt/boot/kernel8.img
cp arch/arm64/boot/dts/broadcom/bcm2710-rpi-3-b.dtb ../mnt/boot/
make ARCH=arm64 CROSS_COMPILE=aarch64-linux-gnu- INSTALL_MOD_PATH=$(dirname $PWD)/mnt modules_install
cd ..



# compress image

umount mnt/boot mnt
zip pi64.zip pi64.img
