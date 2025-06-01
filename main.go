package main

import (
	"flag"
	"fmt"
	"os"
	"sync"
	"path/filepath"
)

// GitVersion is set by the Makefile and contains the version string.
var GitVersion = ""

var stats struct {
	total              int
	errorsNotRegular   int
	errorsOpening      int
	errorsWritingXattr int
	errorsOther        int
	inprogress         int
	corrupt            int
	timechange         int
	outdated           int
	newfile            int
	ok                 int
}

var args struct {
	remove    bool
	recursive bool
	q         bool
	qq        bool
	dryrun    bool
	fix       bool
	migrate   bool
	plaintext bool
	printok   bool // output from <ok> is extended with sha and time info
	mt        int  // multithreading
}

type StatsReport struct {
	path string
	err error
}

var fileChan = make(chan string)
var statsChan = make(chan StatsReport)
var wg sync.WaitGroup
var statsgroup sync.WaitGroup

func consumeStatsForever() {
	defer statsgroup.Done()
	for report := range statsChan {
		recordStat(report.path, report.err)
	}
}

func consumeFilesForever() {
	defer wg.Done()
	for path := range fileChan {
		statsChan <- StatsReport{path: path, err: checkFile(path)}
	}
}

func recordStat(path string, err error) {
	stats.total++
	switch err {
	case ErrCorrupt:
		stats.corrupt++
	case ErrTimeChange:
		stats.timechange++
	case ErrOutdated:
		stats.outdated++
	case ErrNoMetadata:
		stats.newfile++
	case ErrInProgress:
		stats.inprogress++
	case ErrWriteAttr:
		stats.errorsWritingXattr++
	case ErrOsOpen:
		stats.errorsOpening++
	case nil:
		stats.ok++
	}
}

// walkFn is used when `goshatag` is called with the `--recursive` option. It is the function called
// for each file or directory visited whilst traversing the file tree.
func walkFn(path string, info os.FileInfo, err error) error {
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error accessing %q: %v\n", path, err)
		stats.errorsOpening++
	} else if info.Mode().IsRegular() {
		if args.mt != 0 {
			fileChan <- path
		} else {
			err := checkFile(path)
			recordStat(path, err)
		}
	} else if !info.IsDir() {
		if !args.qq {
			fmt.Printf("<nonregular> %s\n", path)
		}
	}
	return nil
}

// processArg is called for each command-line argument given. For regular files it will call
// `checkFile`. Directories will be processed recursively provided the `--recursive` flag is set.
// Symbolic links are not followed.
func processArg(fn string) {
	fi, err := os.Lstat(fn) // Using Lstat to be consistent with filepath.Walk for symbolic links.
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		stats.errorsOpening++
	} else if fi.Mode().IsRegular() {
		if args.mt != 0 {
			fileChan <- fn
		} else {
			err := checkFile(fn)
			recordStat(fn, err)
		}
	} else if fi.IsDir() {
		if args.recursive {
			filepath.Walk(fn, walkFn)
		} else {
			fmt.Fprintf(os.Stderr, "Error: %q is a directory, did you mean to use the '-recursive' option?\n", fn)
			stats.errorsNotRegular++
		}
	} else {
		fmt.Fprintf(os.Stderr, "Error: %q is not a regular file.\n", fn)
		stats.errorsNotRegular++
	}
}

func main() {
	const myname = "goshatag"

	if GitVersion == "" {
		GitVersion = "(version unknown)"
	}

	flag.BoolVar(&args.remove, "remove", false, "Remove any previously stored extended attributes.")
	flag.BoolVar(&args.q, "q", false, "quiet: don't print <ok> files")
	flag.BoolVar(&args.qq, "qq", false, "quiet²: Only print <corrupt> files and errors")
	flag.BoolVar(&args.recursive, "recursive", false, "Recursively descend into subdirectories. "+
		"Symbolic links are not followed.")
	flag.BoolVar(&args.dryrun, "dry-run", false, "don't make any changes")
	flag.BoolVar(&args.fix, "fix", false, "fix the stored sha256 on corrupt files")
	flag.BoolVar(&args.migrate, "migrate", false, "migrate from user.shatag.{sha256,ts} to user.hash")
	flag.BoolVar(&args.plaintext, "plaintext", false, "use user.shatag.{sha256,ts} instead of user.hash")
	flag.BoolVar(&args.printok, "printok", false, "print sha256 and ts for <ok> files")
	flag.IntVar(&args.mt, "mt", 0, "number of threads to read files across (0 preserves ordering, careful going higher than 1 with spinning rust!)")
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "%s %s\n", myname, GitVersion)
		fmt.Fprintf(os.Stderr, "Usage: %s [OPTIONS] FILE [FILE2 ...]\n", myname)
		fmt.Fprintf(os.Stderr, "Options:\n")
		flag.PrintDefaults()
		os.Exit(1)
	}
	flag.Parse()
	if flag.NArg() == 0 {
		flag.Usage()
	}
	if args.qq {
		// quiet2 implies quiet
		args.q = true
	}

	if args.mt > 0 {
		fileChan = make(chan string, args.mt)
		statsChan = make(chan StatsReport, args.mt)
		for range args.mt {
			wg.Add(1)
			fmt.Println("starting new mt thread")
			go consumeFilesForever()
		}
		statsgroup.Add(1)
		go consumeStatsForever()
	}

	for _, fn := range flag.Args() {
		processArg(fn)
	}
	if args.mt > 0 {
		close(fileChan)
		wg.Wait()
		close(statsChan)
		statsgroup.Wait()
	}

	if stats.corrupt > 0 {
		os.Exit(5)
	}

	totalErrors := stats.errorsOpening + stats.errorsNotRegular + stats.errorsWritingXattr +
		stats.errorsOther
		fmt.Printf("Stats: total: %d ok: %d total errors: %d corrupt: %d\n", stats.total, stats.ok, totalErrors, stats.corrupt)
	if totalErrors > 0 {
		if stats.errorsOpening == totalErrors {
			os.Exit(2)
		} else if stats.errorsNotRegular == totalErrors {
			os.Exit(3)
		} else if stats.errorsWritingXattr == totalErrors {
			os.Exit(4)
		}
		os.Exit(6)
	}
	if (stats.ok + stats.outdated + stats.timechange + stats.newfile) == stats.total {
		os.Exit(0)
	}
	os.Exit(6)
}
