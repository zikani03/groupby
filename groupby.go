package main

import (
	"fmt"
	"flag"
	"os"
	"strconv"
	"time"
)

// const SUBDIRECTORY_INNER: &'static str = "├───";
// const SUBDIRECTORY_PIPE: &'static str = "│";
// const SUBDIRECTORY_LINK: &'static str = "└───";

type Tree struct {
	Root *Node
	MaxDepth int
}

type Node struct {
	FileName string
	Year int
	Month time.Month
	Day int
	Next *Node
	Children *Node
}

type TreeVisitor interface {
	Visit(n *Node, depth int)
}

type PrintingVisitor struct {
}

func NewPrintingVisitor() *PrintingVisitor {
	return &PrintingVisitor {}
}

func (p *PrintingVisitor) Visit(n *Node, depth int) {
	if depth == 0 {
		fmt.Println("", n.FileName)
	} else {
		fmt.Println("-- ", n.FileName)
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
	return &Tree{ 
		Root: NewNode(directory, year, month, day),
		MaxDepth: maxDepth,
	}
}

func (t *Tree) Build() {
	file, err := os.Open(t.Root.FileName)
	
	if err != nil {
		panic(err)
	}

	files, err := file.Readdir(-1)

	if err != nil {
		panic(err)
	}

	if files == nil {
		panic(files)
	}

	for _, f := range files {
		t.AddEntry(f)
	}
}

func (t *Tree) AddEntry(file os.FileInfo) {
	year, month, day := GetFileInfoYMD(file)
	var node = NewNode(file.Name(), year, month, day)

	yearStr, monthStr, dayStr := fmt.Sprintf("%d", year), fmt.Sprintf("%d", month), fmt.Sprintf("%d", day)
	var yearNode = t.Root.Search(yearStr)

	// year node
	if (yearNode == nil) {
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

			if (dayNode == nil) {
				dayNode = NewNode(dayStr, year, month, day)
				monthNode.AddChild(dayNode)
			}

			dayNode.AddChild(node)

		} else {
			monthNode.AddChild(node)
		}
	}
}

func (t *Tree) Visit(visitor *PrintingVisitor) {
	t.Root.Visit(visitor, 0)
}

func NewNode(fileName string, year int, month time.Month, day int) *Node {
	return &Node {
		FileName: fileName,
		Year: year,
		Month: month,
		Day: day,
		Next: nil,
		Children: nil,
	}
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

func (n *Node) Visit(visitor *PrintingVisitor, depth int) {
	visitor.Visit(n, depth)
	var cur = n.Children
	for cur != nil {
		cur.Visit(visitor, depth + 1)
		cur = cur.Next
	}
}

/// Usage
///
/// groupby [options] DIRECTORY
///
/// ## Options
///
/// ```
/// -c   --created Group files by the date they were created
/// -m   --modified Group files by the date they were modified
/// -n   --dry-run Show the output of how the files will be grouped
/// -d   --depth N How deep to create the directory hierarchy (maximum: 3)
///                corresponding to 1 - year, 2 - month, 3 - day
/// -D   --group-dirs Move directories into groups as well - by default only
///                   regular files are grouped
/// -R   --recurse Group files in subdirectories
/// -h   --help Show the help information and exit
/// -v   --verbose Show verbose output
///      --version Show the program version
/// ```
///
/// ## Examples
///
/// ```
/// $ groupby -c -D -R -v -d 3 ./my_directory
/// $ groupby --modified -DRv -d 3 ./my_directory
/// ```
///
/// 2016
/// ├─── Jan
/// │    └── 01
/// │        └── my_file.txt
/// └── Feb
///     └── 01
///         └── my_file_2.txt
///
func main() {
	flag.Parse()
	directory := flag.Arg(0)

	depth, err := strconv.Atoi(flag.Arg(1))

	if err != nil {
		panic(err)
	}

	var tree = NewTree(directory, depth)

	tree.Build()

	tree.Visit(NewPrintingVisitor())
}