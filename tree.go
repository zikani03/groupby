package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
)

type Tree struct {
	Root     *Node
	MaxDepth int
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
