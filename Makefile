platform := $(shell uname -s | tr '[:upper:]' '[:lower:]')
arch := $(shell uname -m | tr '[:upper:]' '[:lower:]')

.PHONY: build clean test symbolic

test: build
	go test ./pkg/symbolic

build: symbolic
	go build ./pkg/symbolic

symbolic: include/symbolic.h lib/${platform}_${arch}/libsymbolic_cabi.*

include/symbolic.h: include third_party/symbolic/symbolic-cabi/include/symbolic.h
	cp third_party/symbolic/symbolic-cabi/include/symbolic.h include/

include:
	mkdir -p include

lib/${platform}_${arch}/libsymbolic_cabi.%: lib/${platform}_${arch} third_party/symbolic/target/release/libsymbolic_cabi.*
	cp third_party/symbolic/target/release/libsymbolic_cabi.* lib/${platform}_${arch}/

lib/${platform}_${arch}:
	mkdir -p lib/${platform}_${arch}

symbolic/target/release/libsymbolic_cabi.%:
	$(MAKE) -C third_party/symbolic/symbolic-cabi release
ifeq ($(platform), darwin)
		sudo install_name_tool -id @rpath/libsymbolic_cabi.dylib third_party/symbolic/target/release/libsymbolic_cabi.dylib
endif

clean:
	rm -rf include lib third_party/symbolic/target
