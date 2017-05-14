.PHONY: build validate release

build: build/pi64-lite.zip build/pi64-desktop.zip

build/pi64-lite.zip: build/linux build/userland build/firmware
	pi64-build -build-dir ./build -version lite

build/pi64-desktop.zip: build/linux build/userland build/firmware
	pi64-build -build-dir ./build -version desktop

build/userland:
	bash make/videocore

build/firmware:
	bash make/firmware

build/linux:
	bash make/kernel

validate:
	bash make/validate

release:
	bash make/release
