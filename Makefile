BUILD_VERSION := $(shell cat version)

all: build

build:
	bash .cross_compile.sh

clean:
	rm -rf dist

install:
	bash .cross_compile.sh install

uninstall:
	bash .cross_compile.sh uninstall

completion:
	bash .cross_compile.sh completion

release: build
	gh release create $(BUILD_VERSION) dist/mmh_darwin_* dist/checksums.txt \
		--title $(BUILD_VERSION) --notes "Release $(BUILD_VERSION)"

.PHONY: all build clean install uninstall completion release
