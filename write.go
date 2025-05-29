package main

// This file has all functions that actually change something on disk

import (
	"errors"
	"os"
	"fmt"
	"runtime"

	"github.com/pkg/xattr"
)

// storeAttr stores "attr" into extended attributes.
// used to look like this afterwards:
//
//	$ getfattr -d foo.txt
//	user.shatag.sha256="dc9fe2260fd6748b29532be0ca2750a50f9eca82046b15497f127eba6dda90e8"
//	user.shatag.ts="1748509446.586368096"
//
// now it looks like this
//	$ getfattr -e hex -d foo.txt
//	user.hash=0xdc9fe2260fd6748b29532be0ca2750a50f9eca82046b15497f127eba6dda90e8     313734383530393434362e353836333638303936
func storeAttr(f *os.File, attr fileAttr) (err error) {
	if args.dryrun {
		return nil
	}
	if runtime.GOOS == "darwin" {
		// SMB or MacOS bug: when working on an SMB mounted filesystem on a Mac, it seems the call
		// to `fsetxattr` does not update the xattr but removes it instead. So it takes two runs
		// of `cshatag` to update the attribute.
		// To work around this issue, we remove the xattr explicitly before setting it again.
		// https://github.com/rfjakob/cshatag/issues/8
		removeAttr(f)
	}
	if args.plaintext {
		err = xattr.FSet(f, xattrTs, []byte(attr.ts.prettyPrint()))
		if err != nil {
			return
		}
		err = xattr.FSet(f, xattrSha256, []byte(fmt.Sprintf("%x", attr.sha256)))
	} else {
		packed := make([]byte, 52)
		copy(packed, attr.sha256)
		copy(packed[32:], attr.ts.prettyPrint())
		err = xattr.FSet(f, xattrCombined, packed)
	}
	return
}

func removePlaintextAttr(f *os.File) error {
	err1 := xattr.FRemove(f, xattrTs)
	err2 := xattr.FRemove(f, xattrSha256)
	if err1 != nil && err2 != nil {
		return errors.New(err1.Error() + "  " + err2.Error())
	}
	if err1 != nil {
		return err1
	}
	if err2 != nil {
		return err2
	}
	return nil
}

// removeAttr removes any previously stored extended attributes. Returns an error
// if removal of either the timestamp or checksum xattrs fails.
func removeAttr(f *os.File) error {
	if args.dryrun {
		return nil
	}
	newError := xattr.FRemove(f, xattrCombined)
	if args.migrate || args.plaintext {
		newError = removePlaintextAttr(f)
	}
	return newError
}
