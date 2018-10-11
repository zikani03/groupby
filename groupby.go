package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

const (
	GROUPBY_VERSION    = "0.1.0"
	SUBDIRECTORY_INNER = "├───"
	SUBDIRECTORY_PIPE  = "│"
	SUBDIRECTORY_LINK  = "└───"
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
	includeHidden     bool
	dryRun            bool
	excludePattern    string
	verbose           bool
	version           bool
)

func init() {
	flag.StringVar(&directory, "d", "", "\tDirectory containing files to group")
	flag.StringVar(&outputDirectory, "o", "", "\tDirectory to move grouped files to")
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
	flag.BoolVar(&includeHidden, "a", false, "\tInclude hidden files and directories (starting with .)")
	// flag.String(&exclude, "exclude", "Exclude files or directory matching a specified pattern")
	// flag.BoolVar(&recurse, "R", "recurse" "Group files in subdirectories")
	flag.BoolVar(&verbose, "verbose", false, "\tShow verbose output")
	flag.BoolVar(&verbose, "v", false, "\tShow verbose output")
	flag.BoolVar(&version, "version", false, "\tShow the program version and exit")
}

type GroupbyError struct {
	Message string
}

func groupbyError(msg string) GroupbyError {
	return GroupbyError{
		Message: msg,
	}
}

func (e GroupbyError) Error() string {
	return e.Message
}

type Tree struct {
	Root     *Node
	MaxDepth int
}

type Node struct {
	FileName string
	Year     int
	Month    time.Month
	Day      int
	Next     *Node
	Children *Node
}

type NodeVisitor interface {
	Visit(n *Node, depth int)
}

type PrintingVisitor struct {
	NodeVisitor
	currentLevel  int
	previousLevel int
	indentLevel   int
}

func NewPrintingVisitor() *PrintingVisitor {
	return &PrintingVisitor{}
}

func (p *PrintingVisitor) Visit(n *Node, depth int) {
	p.currentLevel = depth
	if p.currentLevel == 0 {
		p.indentLevel = 0
		p.previousLevel = 0
		fmt.Println(n.FileName)
		return
	}

	if depth >= 2 {
		p.indentLevel = depth
		for i := 0; i < p.indentLevel-1; i++ {
			fmt.Printf("   ")
		}
	}

	if !n.HasNext() {
		if depth == 2 {
			fmt.Println("└──", MonthAsName(n.FileName))
		} else {
			fmt.Println("└──", n.FileName)
		}
	} else {
		if depth == 2 {
			fmt.Println("├──", MonthAsName(n.FileName))
		} else {
			fmt.Println("├──", n.FileName)
		}
	}
	p.previousLevel = depth
}

func MonthAsName(monthStr string) string {
	monthIdx, err := strconv.Atoi(monthStr)
	if err != nil {
		return monthStr
	}

	switch monthIdx {
	case 1:
		return "January"
	case 2:
		return "February"
	case 3:
		return "March"
	case 4:
		return "April"
	case 5:
		return "May"
	case 6:
		return "June"
	case 7:
		return "July"
	case 8:
		return "August"
	case 9:
		return "September"
	case 10:
		return "October"
	case 11:
		return "November"
	case 12:
		return "December"
	}
	return ""
}

// Adapted from: https://stackoverflow.com/a/21067803
// moveOrCopyFile moves or copies a file from src to dst.
// If src and dst files exist, and are the same, then return success.
// Attempt to move the file using os.Rename if the copyOnly flag is false
// Otherwise, we attempt to create a hard link between the two files.
func moveOrCopyFile(src, dst string) (err error) {
	sfi, err := os.Stat(src)
	if err != nil {
		return
	}
	if !sfi.Mode().IsRegular() {
		// cannot copy non-regular files (e.g., directories,
		// symlinks, devices, etc.)
		return fmt.Errorf("moveOrCopyFile: non-regular source file %s (%q)", sfi.Name(), sfi.Mode().String())
	}
	dfi, err := os.Stat(dst)
	if err != nil {
		if !os.IsNotExist(err) {
			return
		}
	} else {
		if !(dfi.Mode().IsRegular()) {
			return fmt.Errorf("moveOrCopyFile: non-regular destination file %s (%q)", dfi.Name(), dfi.Mode().String())
		}
		if os.SameFile(sfi, dfi) {
			return
		}
	}

	// User wants to actually move the files
	if !copyOnly {
		if err = os.Rename(src, dst); err == nil {
			return
		}
	}
	// Creates a hardlink to the source
	if err = os.Link(src, dst); err == nil {
		return
	}
	return err
}

type DirectoryVisitor struct {
	NodeVisitor
	rootDir          string
	flatten          bool
	maxDepth         int
	currentLevel     int
	currentLevelDir  string
	previousLevel    int
	previousLevelDir string
	indentLevel      int
	pathParts        []string
}

func NewDirectoryVisitor(root string, flatten bool, maxDepth int) *DirectoryVisitor {
	return &DirectoryVisitor{
		rootDir:   root,
		flatten:   flatten,
		pathParts: []string{root, "", "", ""},
		maxDepth:  maxDepth,
	}
}

func (v *DirectoryVisitor) Visit(n *Node, depth int) {
	v.currentLevel = depth
	v.currentLevelDir = v.rootDir
	if v.currentLevel == 0 {
		v.indentLevel = 0
		v.previousLevel = 0
		v.pathParts[1] = ""
		v.pathParts[2] = ""
		v.pathParts[3] = ""
		return
	}

	if depth == 2 {
		v.pathParts[depth-1] = MonthAsName(n.FileName)
	} else {
		v.pathParts[depth-1] = n.FileName
	}

	// We're probably at a month
	if depth == 3 && !n.HasNext() {
		v.pathParts[depth] = ""
	}
	dirs := []string{outputDirectory}
	dirs = append(dirs, v.pathParts[:v.maxDepth]...)
	if flatten && n.HasChildren() {
		return
	}
	var dest string
	source := path.Join(v.rootDir, n.FileName)
	destParts := []string{outputDirectory}
	if flatten {
		dirs = []string{v.rootDir, strings.Join(v.pathParts[:v.maxDepth], "-")}
		flattenedParent := strings.Join(v.pathParts[:v.maxDepth], "-")
		destParts = append(destParts, flattenedParent)
		destParts = append(destParts, n.FileName)
	} else {
		destParts = append(destParts, v.pathParts...)
	}
	dest = path.Join(destParts...)
	// Create the destination directories
	perm := os.FileMode(0755)
	rootStat, _ := os.Stat(directory)
	// use permissions of the root directory
	perm = rootStat.Mode()
	err := os.MkdirAll(path.Join(dirs...), perm)
	if err != nil {
		// error
	}
	// Move the file from the source to the directory
	moveOrCopyFile(source, dest)
	v.previousLevel = depth
}

type Visitors struct {
	NodeVisitor
	visitors []NodeVisitor
}

func NewVisitors(visitors ...NodeVisitor) *Visitors {
	return &Visitors{
		visitors: visitors,
	}
}

func (v *Visitors) Visit(n *Node, depth int) {
	for _, visitor := range v.visitors {
		visitor.Visit(n, depth)
	}
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

func NewTree(directory string, maxDepth int) *Tree {
	year, month, day := GetYMD(directory)
	dirPath, err := filepath.Abs(directory)
	if err != nil {
		log.Fatal(err)
	}
	return &Tree{
		Root:     NewNode(dirPath, year, month, day),
		MaxDepth: maxDepth,
	}
}

func (t *Tree) Build() error {
	file, err := os.Open(t.Root.FileName)

	if err != nil {
		return err
	}

	files, err := file.Readdir(-1)

	if err != nil {
		return err
	}

	if files == nil {
		return groupbyError("Directory is empty or cannot be read")
	}

	for _, f := range files {
		if ignoreDirectories && f.IsDir() {
			continue
		} else {
			t.AddEntry(f)
		}
	}
	return nil
}

func (t *Tree) AddEntry(file os.FileInfo) {

	if strings.HasPrefix(file.Name(), ".") && !includeHidden {
		return
	}

	year, month, day := GetFileInfoYMD(file)
	var node = NewNode(file.Name(), year, month, day)

	yearStr, monthStr, dayStr := fmt.Sprintf("%d", year), fmt.Sprintf("%d", month), fmt.Sprintf("%d", day)
	var yearNode = t.Root.Search(yearStr)

	// year node
	if yearNode == nil {
		yearNode = NewNode(yearStr, year, month, day)
		t.Root.AddChild(yearNode)
	}

	if t.MaxDepth == 1 {
		yearNode.AddChild(node)
	}

	if t.MaxDepth >= 2 {
		// month node
		var monthNode = yearNode.Search(monthStr)

		if monthNode == nil {
			monthNode = NewNode(monthStr, year, month, day)
			yearNode.AddChild(monthNode)
		}

		if t.MaxDepth == 3 {
			var dayNode = monthNode.Search(dayStr)

			if dayNode == nil {
				dayNode = NewNode(dayStr, year, month, day)
				monthNode.AddChild(dayNode)
			}

			dayNode.AddChild(node)

		} else {
			monthNode.AddChild(node)
		}
	}
}

func (t *Tree) Visit(visitor NodeVisitor) {
	t.Root.Visit(visitor, 0)
}

func NewNode(fileName string, year int, month time.Month, day int) *Node {
	return &Node{
		FileName: fileName,
		Year:     year,
		Month:    month,
		Day:      day,
		Next:     nil,
		Children: nil,
	}
}

func (n *Node) HasNext() bool {
	return n.Next != nil
}

func (n *Node) HasChildren() bool {
	return n.Children != nil
}

func (n *Node) AddChild(node *Node) {
	var oldNext = n.Children
	n.Children = node
	n.Children.Next = oldNext
}

// Search the children nodes for the value in the string
func (n *Node) Search(value string) *Node {
	var cur = n.Children
	for cur != nil {
		if cur.FileName == value {
			return cur
		}
		cur = cur.Next
	}
	return nil
}

func (n *Node) Visit(visitor NodeVisitor, depth int) {
	visitor.Visit(n, depth)
	var cur = n.Children
	for cur != nil {
		cur.Visit(visitor, depth+1)
		cur = cur.Next
	}
}

func main() {
	flag.Parse()

	if version {
		fmt.Println("groupby ", GROUPBY_VERSION, " - Group files and directories by the date they were created or modified")
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
		fmt.Println("Failed to build directory tree")
		os.Exit(1)
	}
	printingVisitor := NewPrintingVisitor()
	if dryRun {
		tree.Visit(printingVisitor)
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
