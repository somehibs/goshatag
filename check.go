package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"errors"
	"io"
	"os"
	"runtime"
	"strconv"
	"strings"
	"syscall"

	"github.com/pkg/xattr"
)

const xattrSha256 = "user.shatag.sha256"
const xattrTs = "user.shatag.ts"
const xattrCombined = "user.hash"
const zeroSha256 = "0000000000000000000000000000000000000000000000000000000000000000"
var (
	// file sha doesn't match current sha
	ErrCorrupt = errors.New("file_corrupt")
	// stored time doesn't match modified time
	ErrOutdated = errors.New("outdated")
	// time changed so regenerate sha without checking
	ErrTimeChange = errors.New("time_changed")
	// no time or sha so regenerate sha
	ErrNoMetadata = errors.New("no_metadata")
	// file was modified during sha computation
	ErrInProgress = errors.New("in_progress")
	// file was modified during sha computation
	ErrWriteAttr = errors.New("write_attr_failed")
	// can't remove attr in -remove mode, can't fetch actual modified time or sha256 in regular mode
	ErrOther = errors.New("other")
	// can't open file
	ErrOsOpen = errors.New("os_open")
)

type fileTimestamp struct {
	s  uint64
	ns uint32
}

func (ts *fileTimestamp) prettyPrint() string {
	return fmt.Sprintf("%010d.%09d", ts.s, ts.ns)
}

// equalTruncatedTimestamp compares ts and ts2 with 100ns resolution (Linux) or 1s (MacOS).
// Why 100ns? That's what Samba and the Linux SMB client supports.
// Why 1s? That's what the MacOS SMB client supports.
func (ts *fileTimestamp) equalTruncatedTimestamp(ts2 *fileTimestamp) bool {
	if ts.s != ts2.s {
		return false
	}
	// We only look at integer seconds on MacOS, so we are done here.
	if runtime.GOOS == "darwin" {
		return true
	}
	if ts.ns/100 != ts2.ns/100 {
		return false
	}
	return true
}

type fileAttr struct {
	ts          fileTimestamp
	sha256      []byte
	storageType int
}

func (a *fileAttr) shaPrint() (sha string) {
	if a.sha256 == nil {
		sha = zeroSha256
	} else {
		sha = fmt.Sprintf("%x", a.sha256)
	}
	return
}
func (a *fileAttr) prettyPrint() string {
	sha := a.shaPrint()
	return fmt.Sprintf("%s %s", sha, a.ts.prettyPrint())
}

func parseStoredTs(storedTs []byte) (ts fileTimestamp) {
	parts := strings.SplitN(string(storedTs), ".", 2)
	ts.s, _ = strconv.ParseUint(parts[0], 10, 64)
	if len(parts) > 1 {
		ns64, _ := strconv.ParseUint(parts[1], 10, 32)
		ts.ns = uint32(ns64)
	}
	return
}

// getStoredAttr reads the stored extended attributes from a file. The file
// used to look like this:
//
//	$ getfattr -d foo.txt
//	user.shatag.sha256="dc9fe2260fd6748b29532be0ca2750a50f9eca82046b15497f127eba6dda90e8"
//	user.shatag.ts="1748509446.586368096"
//
// now it looks more like this
//	$ getfattr -e hex -d foo.txt
//	user.hash=0xdc9fe2260fd6748b29532be0ca2750a50f9eca82046b15497f127eba6dda90e8     313734383530393434362e353836333638303936
// 
var newStorage = errors.New("new_storage")
func getStoredAttr(f *os.File) (attr fileAttr, err error) {
	attr.sha256 = nil
	if args.migrate || args.plaintext {
		val, err := xattr.FGet(f, xattrSha256)
		if err == nil {
			unhexed, err := hex.DecodeString(string(val))
			if err != nil {
				return attr, err
			}
			attr.sha256 = make([]byte, 32)
			copy(attr.sha256, unhexed)
		}
		val, err = xattr.FGet(f, xattrTs)
		if err == nil {
			attr.ts = parseStoredTs(val)
			attr.storageType = 1
		}
	}
	// always try to get the new storage, avoiding excess writes during migration
	val, err := xattr.FGet(f, xattrCombined)
	if err == nil {
		attr.sha256 = make([]byte, 32)
		copy(attr.sha256, val[:32])
		attr.ts = parseStoredTs(val[32:])
		attr.storageType = 2
	}
	return attr, nil
}

// getMtime reads the actual modification time of file "f" from disk.
func getMtime(f *os.File) (ts fileTimestamp, err error) {
	fi, err := f.Stat()
	if err != nil {
		return
	}
	ts.s = uint64(fi.ModTime().Unix())
	ts.ns = uint32(fi.ModTime().Nanosecond())
	return
}

// getActualAttr reads the actual modification time and hashes the file content.
func getActualAttr(f *os.File) (attr fileAttr, err error) {
	attr.ts, err = getMtime(f)
	if err != nil {
		return attr, err
	}
	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return attr, err
	}
	// Check if the file was modified while we were computing the hash
	ts2, err := getMtime(f)
	if err != nil {
		return attr, err
	} else if attr.ts != ts2 {
		return attr, syscall.EINPROGRESS
	}
	attr.sha256 = h.Sum(nil)
	return attr, nil
}

// printComparison prints something like this:
//
//	stored: faa28bfa6332264571f28b4131b0673f0d55a31a2ccf5c873c435c235647bf76 1560177189.769244818
//	actual: dc9fe2260fd6748b29532be0ca2750a50f9eca82046b15497f127eba6dda90e8 1560177334.020775051
func printComparison(stored fileAttr, actual fileAttr) {
	fmt.Printf(" stored: %s\n actual: %s\n", stored.prettyPrint(), actual.prettyPrint())
}

func checkFile(fn string) (err error) {
	f, err := os.Open(fn)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err)
		return ErrOsOpen
	}
	defer f.Close()

	if args.remove {
		if err = removeAttr(f); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %s\n", err)
			return ErrOther
		}
		if !args.q {
			fmt.Printf("<removed xattr> %s\n", fn)
		}
		return
	}

	stored, _ := getStoredAttr(f)
	actual, err := getActualAttr(f)
	if err == syscall.EINPROGRESS {
		if !args.qq {
			fmt.Printf("<concurrent modification> %s\n", fn)
		}
		return ErrInProgress
	} else if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err)
		return ErrOther
	}

	var allZeroTimeStamp fileTimestamp

	// files can only be verified / corrupt if they have actual storage, otherwise they're just new
	if stored.storageType != 0 && stored.ts.equalTruncatedTimestamp(&actual.ts) {
		if bytes.Equal(stored.sha256, actual.sha256) {
			if !args.q {
				if args.printok {
					fmt.Printf("%s  %s\n", stored.shaPrint(), fn)
				} else {
					fmt.Printf("<ok> %s\n", fn)
				}
			}
			if !args.migrate {
				return
			}
		} else {
			fixing := " Keeping hash as-is (use -fix to force hash update)."
			if args.fix {
				fixing = " Fixing hash (-fix was passed)."
			}
			fmt.Fprintf(os.Stderr, "Error: corrupt file %q. %s\n", fn, fixing)
			fmt.Printf("<corrupt> %s\n", fn)
			err = ErrCorrupt
		}
	} else if bytes.Equal(stored.sha256, actual.sha256) {
		if !args.qq {
			fmt.Printf("<timechange> %s\n", fn)
		}
		err = ErrTimeChange
	} else if stored.sha256 == nil && (stored.ts == allZeroTimeStamp) {
		// no metadata indicates a 'new' file
		if !args.qq {
			fmt.Printf("<new> %s\n", fn)
		}
		err = ErrNoMetadata
	} else {
		// timestamp is outdated
		if !args.qq {
			fmt.Printf("<outdated> %s\n", fn)
		}
		err = ErrOutdated
	}

	if !args.qq {
		printComparison(stored, actual)
	}

	// Only update the stored attribute if it is not corrupted **OR**
	// if argument '-fix' been given **OR**
	// if it hasn't been written and the file has a modified timestamp of 0
	if stored.storageType == 0 || stored.ts != actual.ts || args.fix || (args.migrate && stored.storageType != 2) {
		if args.migrate {
			// don't allow the old attributes to exist
			removePlaintextAttr(f)
		}
		err = storeAttr(f, actual)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %s\n", err)
			return ErrWriteAttr
		}
	}
	return
}
