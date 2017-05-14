.PHONY: build validate release

build: build/pi64-lite.zip build/pi64-desktop.zip

build/pi64-lite.zip: build/linux build/userland build/firmware
	bash make/image lite

build/pi64-desktop.zip: build/linux build/userland build/firmware
	bash make/image desktop

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
