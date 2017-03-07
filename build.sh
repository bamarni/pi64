#!/bin/sh

set -ex

# dependencies

apt-get update
apt-get install -y bc build-essential gcc-aarch64-linux-gnu git unzip qemu-user-static multistrap zip


mkdir -p build
cd build



# build kernel

git clone --depth=1 -b rpi-4.9.y https://github.com/raspberrypi/linux.git

cd linux
make ARCH=arm64 CROSS_COMPILE=aarch64-linux-gnu- bcmrpi3_defconfig
echo "CONFIG_KEYS_COMPAT=y" >> .config
make -j 3 ARCH=arm64 CROSS_COMPILE=aarch64-linux-gnu-
cd ..



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

mkdosfs -n boot -F 32 -v $boot_dev
mkfs.ext4 -O ^huge_file $root_dev



# build rootfs

mkdir -p mnt
mount -v $root_dev mnt -t ext4

mkdir -p mnt/dev
mount -o bind /dev mnt/dev

multistrap -a arm64 -d $PWD/mnt -f ../multistrap.conf

cp /usr/bin/qemu-aarch64-static mnt/usr/bin/qemu-aarch64-static

cat << EOF | chroot mnt
export DEBIAN_FRONTEND=noninteractive DEBCONF_NONINTERACTIVE_SEEN=true
export LC_ALL=C LANGUAGE=C LANG=C

cat > /etc/fstab <<EOL
proc            /proc           proc    defaults          0       0
/dev/mmcblk0p1  /boot           vfat    defaults          0       2
/dev/mmcblk0p2  /               ext4    defaults,noatime  0       1
EOL

mount -t proc proc /proc
mount -t sysfs sys /sys

echo exit 101 > /usr/sbin/policy-rc.d
chmod +x /usr/sbin/policy-rc.d

/var/lib/dpkg/info/dash.preinst install
dpkg --configure -a

rm /usr/sbin/policy-rc.d

echo raspberrypi > /etc/hostname

echo 127.0.1.1 raspberrypi >> /etc/hosts

cat >> /etc/network/interfaces <<EOL
auto eth0
iface eth0 inet dhcp
EOL

useradd -s /bin/bash --create-home -p $(perl -e 'print crypt("raspberry", "password")') pi
echo "pi ALL=(ALL) NOPASSWD: ALL" > /etc/sudoers.d/010_pi-nopasswd

cat > /root/init_setup.sh <<EOL
#!/bin/sh

mount -t proc proc /proc
mount -t sysfs sys /sys
mount /boot
mount -o remount,rw /

parted /dev/mmcblk0 u s resizepart 2 \\\$(expr \\\$(cat /sys/block/mmcblk0/size) - 1)
resize2fs /dev/mmcblk0p2

sed -i 's| init=/root/init_setup.sh||' /boot/cmdline.txt
sync

echo 1 > /proc/sys/kernel/sysrq
rm /root/init_setup.sh
echo b > /proc/sysrq-trigger
EOL

chmod +x /root/init_setup.sh

apt-get clean
rm -rf /var/lib/apt/lists/*
EOF

rm mnt/usr/bin/qemu-aarch64-static



# install boot stuff

mkdir -p mnt/boot
mount -v $boot_dev mnt/boot -t vfat

cp -r ../boot/* mnt/boot

cd linux
cp arch/arm64/boot/Image ../mnt/boot/kernel8.img
cp arch/arm64/boot/dts/broadcom/bcm2710-rpi-3-b.dtb ../mnt/boot/
make ARCH=arm64 CROSS_COMPILE=aarch64-linux-gnu- INSTALL_MOD_PATH=$(dirname $PWD)/mnt modules_install
cd ..



# compress image

umount mnt/boot mnt/dev mnt/proc mnt/sys mnt
zip pi64.img.zip pi64.img
