# Symbolic-Go

Golang bindings for the C-ABI that is produced from the https://github.com/honeycombio/symbolic library. The C-ABI specifically lives in https://github.com/honeycombio/symbolic/tree/master/symbolic-cabi.

We do not support the full API provided by this library. We currently make the following C calls.

* symbolic_init
* symbolic_error_clear
* symbolic_sourcemapcache_from_bytes
* symbolic_sourcemapcache_lookup_token
* symbolic_sourcemapcache_token_match_free
* symbolic_sourcemapcache_free
* symbolic_err_get_last_code
* symbolic_err_get_last_message
* symbolic_err_get_backtrace
* symbolic_str_free

## Releasing

- create a branch named: release/vX.X.X
- CI will build the various libraries for macos/arm64, linux/arm64, linux/x86 and commit them to the branch
- Once the libraries have committed, tag the branch with the version number eg. vX.X.X
