Changelog
---------

*Short changelog - for all the details look at the git log.*

v3.0.0, 2025-05-29
* **Behaviour change: Replace 'user.shatag.ts' and 'user.shatag.sha256' with 'user.hash'**
  user.hash combines sha256 and ts into a single binary encoded value
* New flag: '-migrate' will check for and delete plaintext keys, upgrading in place while checking for errors
* New flag: '-plaintext' will allow you to keep using plaintext keys (512byte inodes, etc)
* ext4 claims that 256 byte inodes have 100 bytes spare for extended attributes...
  but we're actually limited to 68 bytes in a single short key or 48 bytes when using two short keys
* Running cshatag on ext4 with 256 byte inodes and 468,841 files resulted in 1,166,816K of wasted Extended Attributes
* If you want to use goshatag as-is on ext4, it seems that 512 byte inodes would be useful

v2.2.1, 2024-08-23
* Fix `Makefile` to ensure the correct version string is baked
  into the binary ([#29](https://github.com/rfjakob/cshatag/issues/29))

v2.2.0, 2024-08-23
* Do not update corrupt files unless `-fix` is passed ([#28](https://github.com/rfjakob/cshatag/pull/28),
  **behavoir change**)
* Linux: use 100ns resolution when comparing timestamps instead of 1ns
  to match SMB protocol restrictions
  ([#21](https://github.com/rfjakob/cshatag/issues/21)
  [commit](https://github.com/rfjakob/cshatag/commit/3e1f62b38b493b2be75437c208ae7b1d6a90c8e8))
* MacOS: use 1s resolution when comparing timestamps to match
  MacOS SMB client restrictions ([#21](https://github.com/rfjakob/cshatag/issues/21))

v2.1.0, 2022-10-22
* Add `-dry-run` [#22](https://github.com/rfjakob/cshatag/issues/22)
* This version is called `v2.1.0` as opposed to `v2.1` to conform
  to go.mod versioning rules (three-digit semver).

v2.0, 2020-11-15
* Rewrite cshatag in Go
* add MacOS support
* Add `-remove` flag
* Add `-q` and `-qq` flags
* Accept multiple files per invocation to improve performance
* Work around problems on MacOS SMB mounts
  ([#11](https://github.com/rfjakob/cshatag/pull/11))

v1.1, 2019-06-09
* Add test suite (`make test`)
  ([commit](https://github.com/rfjakob/cshatag/commit/74496854e5c934b6809e816b9e854c5c6585a0f4))
* Add Travis CI
* Drop useless trailing null byte from `user.shatag.sha256`

v1.0, 2019-01-02
* Add `make format` target

2019-02-01
* Fix missing null termination in ts buffer that could lead
  to false positives
  ([commit](https://github.com/rfjakob/cshatag/commit/26873dd71656730d5744efb7fa595d529b3c9ae6))

2017-05-04
* Respect `PREFIX` for `make install`
  ([commit](https://github.com/rfjakob/cshatag/commit/8d1225aabb7bdd3750f161133931b1c456bc2fdb))

2016-09-17
* Check for malloc returning NULL
  ([commit](https://github.com/rfjakob/cshatag/commit/ecadbddffb5e23811a9ae4a5265c287d5ae5c151))

2012-12-05
* C source code & man page published on Github
  ([commit](https://github.com/rfjakob/cshatag/commit/5ce7674ea3210fd0bb6b06a81ca8823e0664761a))
