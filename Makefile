.PHONY: all validate release

all: build/pi64-lite.zip build/pi64-desktop.zip

build/pi64-lite.zip: build/pi64-lite.img
	zip -9 -j build/pi64-lite.zip build/pi64-lite.img

build/pi64-desktop.zip: build/pi64-desktop.img
	zip -9 -j build/pi64-desktop.zip build/pi64-desktop.img

build/pi64-lite.img: build/linux build/userland build/firmware
	pi64-build -build-dir ./build -version lite

build/pi64-desktop.img: build/linux build/userland build/firmware
	pi64-build -build-dir ./build -version desktop

build/linux.tar.gz.sig: build/linux.tar.gz
	cd build && gpg2 --output linux.tar.gz.sig --detach-sign linux.tar.gz

build/linux.tar.gz: build/linux
	cd build/linux && tar -zcvf ../linux.tar.gz .

build/linux: build/linux-src build/firmware build/userland
	bash make/linux
	touch build/linux # otherwise make will rebuild that target everytime (as build/linux-src gets altered by make/linux)

build/linux-src:
	bash make/linux-src

build/userland:
	bash make/videocore

build/firmware:
	bash make/firmware

validate:
	bash make/validate

release:
	bash make/release
