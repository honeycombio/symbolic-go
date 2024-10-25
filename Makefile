platform := $(shell uname -s | tr '[:upper:]' '[:lower:]')
arch := $(shell uname -m | tr '[:upper:]' '[:lower:]')

test: build
	go test .

build: include/symbolic.h lib/${platform}_${arch}/libsymbolic_cabi.*
	go build .

include/symbolic.h: include symbolic/symbolic-cabi/include/symbolic.h
	cp symbolic/symbolic-cabi/include/symbolic.h include/

include:
	mkdir -p include

lib/${platform}_${arch}/libsymbolic_cabi.*: lib/${platform}_${arch} symbolic/target/release/libsymbolic_cabi.*
	cp symbolic/target/release/libsymbolic_cabi.* lib/${platform}_${arch}/

lib/${platform}_${arch}:
	mkdir -p lib/${platform}_${arch}

symbolic/target/release/libsymbolic_cabi.%:
	$(MAKE) -C symbolic/symbolic-cabi release
	ifeq ($(platform),darwin)
		sudo install_name_tool -id @rpath/libsymbolic_cabi.dylib symbolic/target/release/libsymbolic_cabi.dylib
	endif

clean:
	rm -rf include lib
