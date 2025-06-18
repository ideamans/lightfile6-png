PWD := $(shell pwd)
UNAME_S := $(shell uname -s)

IMAGEQUANT=libimagequant
IMAGEQUANT_SYS=$(IMAGEQUANT)/imagequant-sys
LIBIMAGEQUANT=$(IMAGEQUANT_SYS)/libimagequant.a

all: $(LIBIMAGEQUANT)

$(IMAGEQUANT_SYS)/Makefile:
	git submodule update --init $(IMAGEQUANT)

$(LIBIMAGEQUANT): $(IMAGEQUANT_SYS)/Makefile
	cd $(IMAGEQUANT_SYS) && make

# iqaにもmake cleanがあるが、releaseは削除しないので直接rmする
.PHONY: clean
clean:
	HERE=$(PWD)
	cd $(IMAGEQUANT_SYS) && make clean
	go clean -cache -testcache

.PHPONY: test
test: $(LIBIMAGEQUANT)
	go test -v ./...