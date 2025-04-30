# Symbolic-Go

Golang bindings for the C-ABI that is produced from the https://github.com/honeycombio/symbolic library. The C-ABI specifically lives in https://github.com/honeycombio/symbolic/tree/master/symbolic-cabi.

We do not support the full API provided by this library. We currently make the following C calls.

* symbolic_err_get_backtrace
* symbolic_err_get_last_code
* symbolic_err_get_last_message
* symbolic_error_clear
* symbolic_init
* symbolic_proguardmapper_free
* symbolic_proguardmapper_get_uuid
* symbolic_proguardmapper_has_line_info
* symbolic_proguardmapper_open
* symbolic_proguardmapper_remap_class
* symbolic_proguardmapper_remap_frame
* symbolic_proguardmapper_remap_method
* symbolic_sourcemapcache_free
* symbolic_sourcemapcache_from_bytes
* symbolic_sourcemapcache_lookup_token
* symbolic_sourcemapcache_token_match_free
* symbolic_str_free
* symbolic_archive_open
* symbolic_archive_from_bytes
* symbolic_archive_free
* symbolic_archive_object_count
* symbolic_archive_get_object
* symbolic_object_free
* symbolic_object_get_arch
* symbolic_object_get_code_id
* symbolic_object_get_debug_id
* symbolic_object_get_kind
* symbolic_object_get_file_format
* symbolic_object_get_features
* symbolic_symcache_from_object
* symbolic_symcache_free
* symbolic_symcache_get_arch
* symbolic_symcache_get_debug_id
* symbolic_symcache_get_version
* symbolic_symcache_lookup

## Developing

### First Time Setup
- Ensure you have the git submodules checked out (`git submodule init && git submodule update`)
- Ensure you have latest stable Rust installed
- Run `make build` in the root of the repo (this builds the `symbolic` package, C ABI, etc.)
- Develop in Go as normal

## Releasing

- create a branch named: release/vX.X.X
- CI will build the various libraries for macos/arm64, linux/arm64, linux/x86 and commit them to the branch
- Once the libraries have committed, tag the branch with the version number eg. vX.X.X
