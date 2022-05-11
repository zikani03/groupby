package main

// Currently 19.1% test coverage

import (
	"flag"
	"fmt"
	"os"
	"strconv"
	"time"
)

const (
	SubdirectoryInner = "├──"
	SubdirectoryPipe  = "│"
	SubdirectoryLink  = "└──"
)

var (
	directory         string
	outputDirectory   string
	copyOnly          bool
	ignoreDirectories bool
	created           bool
	modified          bool
	depth             int = 1
	year              bool
	month             bool
	day               bool
	flatten           bool
	expandMonth       bool
	includeHidden     bool
	dryRun            bool
	excludePattern    string
	filterPattern     string = ""
	verbose           bool
	showVersion       bool
	version           string = "0.0.0"
)

func init() {
	flag.StringVar(&directory, "d", "", "\tDirectory containing files to group")
	flag.StringVar(&outputDirectory, "o", "", "\tDirectory to move grouped files to")
	flag.StringVar(&filterPattern, "e", "", "\tOnly group files matching the given pattern")
	flag.StringVar(&filterPattern, "pattern", "", "\tOnly group files matching the given pattern")
	flag.BoolVar(&copyOnly, "copy-only", false, "\tOnly copy files, do not move them")
	flag.BoolVar(&ignoreDirectories, "ignore-directories", false, "\tIgnore directories and only group files")
	flag.BoolVar(&created, "created", false, "\tGroup files by the date they were created")
	flag.BoolVar(&modified, "modified", true, "\tGroup files by the date they were modified")
	flag.BoolVar(&year, "year", false, "\tGroup by year only")
	flag.BoolVar(&month, "month", false, "\tGroup by year, and then month")
	flag.BoolVar(&day, "day", false, "\tGroup by year, month and then day")
	flag.BoolVar(&flatten, "flatten", false, "\tFlatten the created directory tree folders")
	flag.BoolVar(&dryRun, "dry-run", false, "\tOnly show the output of how the files will be grouped")
	flag.BoolVar(&dryRun, "preview", false, "\tOnly show the output of how the files will be grouped")
	flag.BoolVar(&dryRun, "p", false, "\tOnly show the output of how the files will be grouped (shorthand)")
	flag.BoolVar(&expandMonth, "expand-month", true, "\tUse the English name of the month (e.g. March) instead of the numeric value (default true)")
	flag.BoolVar(&includeHidden, "a", false, "\tInclude hidden files and directories (starting with .)")
	// flag.String(&exclude, "exclude", "Exclude files or directory matching a specified pattern")
	// flag.BoolVar(&recurse, "R", "recurse" "Group files in subdirectories")
	flag.BoolVar(&verbose, "verbose", false, "\tShow verbose output")
	flag.BoolVar(&verbose, "v", false, "\tShow verbose output")
	flag.BoolVar(&showVersion, "version", false, "\tShow the program version and exit")
}

// MonthAsName returns the full month name for the provided monthStr
//
// monthStr is a string usually containing the numeric representation of a
// month (with January=1, February=2, etc.)
//
// If monthStr cannot be casted to an int, returns the provided parameter. If
// monthStr is cast to an int that's not in the range [1, 12] inclusive,
// returns an empty string
func MonthAsName(monthStr string) string {
	monthIdx, err := strconv.Atoi(monthStr)
	if err != nil {
		return monthStr
	}

	if monthIdx < 1 || monthIdx > 12 {
		return ""
	}

	return time.Month(monthIdx).String()
}

// FileNameByDepth returns the filename, potentially modified depending on
// the provided depth
//
// filename is a string containing the name of the file
//
// Depth is how deep down the file structure this file will be. The second
// level is mapped to the month, so the name may be updated to its string representation.
func FileNameByDepth(filename string, depth int) string {
	if depth != 2 || expandMonth == false {
		return filename
	}

	return MonthAsName(filename)
}

// Adapted from: https://stackoverflow.com/a/21067803
// moveOrCopyFile moves or copies a file from src to dst.
// If src and dst files exist, and are the same, then return success.
// Attempt to move the file using os.Rename if the copyOnly flag is false
// Otherwise, we attempt to create a hard link between the two files.
func moveOrCopyFile(src, dst string) (err error) {
	if verbose {
		fmt.Println("Moving from=", src, " to=", dst)
	}
	sfi, err := os.Stat(src)
	if err != nil {
		return err
	}
	if ignoreDirectories && sfi.Mode().IsDir() {
		return
	}
	dfi, err := os.Stat(dst)
	if err != nil {
		if !os.IsNotExist(err) {
			return err
		}
	} else {
		if os.SameFile(sfi, dfi) {
			return
		}
	}
	// User wants to actually move the files
	if !copyOnly {
		if err = os.Rename(src, dst); err == nil {
			return
		}
		return err
	}
	// User wants to -copy-only the files/directories
	// Creates a hardlink to the source
	if err = os.Link(src, dst); err == nil {
		return
	}
	return err
}

func GetYMD(fileName string) (int, time.Month, int) {
	var stat, err = os.Stat(fileName)

	if err != nil {
		// raise error here
		panic(err)
	}
	var tm = stat.ModTime()
	return tm.Year(), tm.Month(), tm.Day()
}

func GetFileInfoYMD(fileInfo os.FileInfo) (int, time.Month, int) {
	var tm = fileInfo.ModTime()
	return tm.Year(), tm.Month(), tm.Day()
}

func main() {
	flag.Parse()

	if showVersion {
		fmt.Println("groupby ", version, " - Group files and directories by the date they were created or modified")
		fmt.Println("By Zikani Nyirenda Mwase ")
		os.Exit(0)
	}

	if _, err := os.Stat(directory); err != nil {
		flag.PrintDefaults()
		os.Exit(0)
	}

	if outputDirectory == "" {
		outputDirectory = directory
	}

	// Build the tree using the deepest depth argument
	if day {
		depth = 3
	} else if month {
		depth = 2
	} else if year {
		depth = 1
	}

	// TODO: Add argument to tree constructor for which file time to use
	var tree = NewTree(directory, depth)
	err := tree.Build()
	if err != nil {
		fmt.Printf("Error: %s", err)
		os.Exit(-1)
	}
	printingVisitor := NewPrintingVisitor()
	if dryRun {
		tree.Visit(printingVisitor)
		fmt.Printf("\n%d directories, %d files\n", tree.Directories(), tree.Files())
		os.Exit(-1)
		return
	}
	directoryVisitor := NewDirectoryVisitor(directory, flatten, depth)
	multiVisitor := NewVisitors(printingVisitor, directoryVisitor)
	if verbose {
		tree.Visit(multiVisitor)
	} else {
		tree.Visit(directoryVisitor)
	}
}
