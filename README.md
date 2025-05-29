[View Changelog](#Changelog)

goshatag is a tool to detect silent data corruption. It is meant to run periodically
and stores the SHA256 of each file as an extended attribute. The project started
as a minimal and fast reimplementation of [shatag](https://github.com/maugier/shatag),
written in Python by Maxime Augier.

goshatag is incompatible with cshatag by default, allowing it to be compatible with 
ext4 in-inode storage. Migration from cshatag is supported with the argument -migration

goshatag remains backwards compatible with cshatag by using the argument -plaintext

See the [Man Page](#man-page) further down this page for details.

Similar Tools
-------------

Checksums stored in extended attributes for each file
* https://github.com/maugier/shatag (the original shatag tool, written in Python)

Checksums stored in single central database
* https://github.com/ambv/bitrot
* https://sourceforge.net/p/yabitrot/code/ci/master/tree/

Checksums stored in one database per directory
* https://github.com/laktak/chkbit-py

Compile
----------------
Needs git, Go and make.

```
$ git clone https://github.com/rfjakob/cshatag.git
$ cd cshatag
$ make
```

Man Page
--------

```
GOSHATAG(1)                      User Manuals                      GOSHATAG(1)

NAME
       goshatag - tag files with sha256 hashes to detect bitrot

SYNOPSIS
       goshatag [OPTIONS] FILE [FILE...]

DESCRIPTION
       goshatag is a minimal and fast re-implementation of shatag
       (  https://github.com/maugier/shatag  ,  written  in  Python  by Maxime
       Augier)
       in a compiled language (since v2.0: Go, earlier versions: C).

       goshatag is a tool to detect silent  data  corruption.  It  writes  the
       mtime  and  the  sha256 checksum of a file into the file's extended at‐
       tributes. The filesystem needs to be mounted  with  user_xattr  enabled
       for  this to work.  When run again, it compares stored mtime and check‐
       sum. If it finds that the mtime  is  unchanged  but  the  checksum  has
       changed,  it  warns  on stderr.  In any case, the status of the file is
       printed to stdout and the stored checksum is updated.

       File statuses that appear on stdout are:
            <new>         file is missing both attributes
            <outdated>    both mtime and checksum have changed
            <ok>          both checksum and mtime stayed the same
            <timechange>  only mtime has changed, checksum stayed the same
            <corrupt>     mtime stayed the same but checksum changed

       goshatag no longer aims to be format-compatible with (c)shatag.  If you
       want  to  use  a  cshatag filesystem with goshatag, please run goshatag
       with the argument -migrate to completion at least once.  The  migration
       process can be interrupted safely.  (see the COMPATIBILITY section).

       cshatag  was written in C in 2012 and has been rewritten in Go in 2019.
       goshatag was forked in 2025.

OPTIONS
       -dry-run    don't make any changes
       -recursive  recursively process the contents of directories
       -remove     remove cshatag's xattrs from FILE
       -q          quiet mode - don't report <ok> files
       -qq         quiet2 mode - only report <corrupt> files and errors
       -fix        fix the stored sha256 on corrupt files
       -migrate    migrate from  user.shatag.{sha256,ts}  to  user.hash  (ext4
       compatibility)
       -plaintext   use  user.shatag.{sha256,ts} instead of user.hash (cshatag
       compatibility)
       -printok    print sha256 and ts for <ok> files

EXAMPLES
       Check all regular files in the current working directory's file tree:
       # goshatag -qq -recursive .
       Errors like corrupt files will be printed to stderr.  Run without "-qq"
       to see progress output.

       To remove extended attributes from all files:
       # goshatag -recursive -remove .

       To migrate from cshatag to goshatag:
       # goshatag -qq -recursive -migrate .

       To  remove  extended attributes from all files, including (c)shatag at‐
       tributes:
       # goshatag -recursive -plaintext -remove .

RETURN VALUE
       0 Success
       1 Wrong number of arguments
       2 One or more files could not be opened
       3 One or more files is not a regular file
       4 Extended attributes could not be written to one or more files
       5 At least one file was found to be corrupt
       6 More than one type of error occurred

COMPATIBILITY
       goshatag writes the extended attribute user.hash with both  the  sha256
       hash  and  the  ASCII  encoded  time,  whereas the original cshatag and
       shatag use the attribut  names  user.shatag.sha256  and  user.shatag.ts
       goshatag and cshatag use higher precision timestamps as compared to the
       shatag (python) version

AUTHORS
       somehibs <@github>
       https://github.com/somehibs/goshatag

       Jakob Unterwurzacher <jakobunt@gmail.com>
       https://github.com/rfjakob/cshatag

COPYRIGHT
       Copyright 2025 somehibs. MIT  License.   Copyright  2012  Jakob  Unter‐
       wurzacher. MIT License.

SEE ALSO
       cshatag(1), shatag(1), sha256sum(1), getfattr(1), setfattr(1)

Version 3.0.0                      MAY 2012                        GOSHATAG(1)
```
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
