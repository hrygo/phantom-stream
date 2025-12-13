.PHONY: all build install test clean help

all: build

help:
	@echo "Phantom Stream Root Makefile"
	@echo "Delegates commands to defender/Makefile"
	@echo ""
	@echo "Available commands:"
	@echo "  make build    - Build the defender binary"
	@echo "  make install  - Install the defender binary to system (may require sudo)"
	@echo "  make test     - Run defender tests"
	@echo "  make clean    - Clean defender artifacts"

build:
	$(MAKE) -C defender build

install:
	$(MAKE) -C defender install

test:
	$(MAKE) -C defender test

clean:
	$(MAKE) -C defender clean
