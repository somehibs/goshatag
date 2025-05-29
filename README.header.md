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
