platform := $(shell uname -s | tr '[:upper:]' '[:lower:]')
arch := $(shell uname -m | tr '[:upper:]' '[:lower:]')

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

symbolic/target/release/libsymbolic_cabi.*:
	$(MAKE) -C symbolic/symbolic-cabi release

clean:
	rm -rf include lib
