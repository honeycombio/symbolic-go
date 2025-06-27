# symbolic-go changelog

## Unreleased
- fix: Fix demangling not working

## 0.0.7
### Enhancements
- feat: expose Swift demangling on dsym

### Maintenance
- Fix segfaults when calling `Symcache.Lookup()`

## 0.0.6
### Maintenance
- Fix build error in 0.0.5 and 0.0.4

## 0.0.5
### Maintenance
- Fix build error in 0.0.4

## 0.0.4
### Enhancements
- update dsym support: expose `SymCaches` field on dSYM `Archive`

## 0.0.3
### Enhancements
- add dsym support
- add proguard support

### Maintenance
- use forks of `symbolic` and `source-map-tests` submodules instead of upstream

## 0.0.2
- initial release
