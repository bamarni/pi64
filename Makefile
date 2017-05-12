.PHONY: build release

build: build/pi64-lite.zip build/pi64-desktop.zip

build/pi64-lite.zip: build/linux build/userland
	bash make/image lite

build/pi64-desktop.zip: build/linux build/userland
	bash make/image desktop

build/userland:
	bash make/videocore

build/linux:
	bash make/kernel

release:
	bash make/release
