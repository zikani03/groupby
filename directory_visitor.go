package main

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"
)

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

	v.pathParts[depth-1] = FileNameByDepth(n.FileName, depth)

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
	sfi, err := os.Stat(source)
	if os.IsNotExist(err) {
		// some internal nodes in our tree won't exist
		return
	}

	destParts := []string{outputDirectory}
	if copyOnly && sfi.IsDir() {
		destParts = []string{}
	}

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
	err = os.MkdirAll(path.Join(dirs...), perm)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create directory %s \n", dirs)
		os.Exit(1)
		return
	}

	if copyOnly && sfi.IsDir() {
		odabs, _ := filepath.Abs(outputDirectory)
		dest = path.Join(odabs, dest)
		source, _ = filepath.Abs(source)
		// fmt.Fprintf(os.Stderr, "Creating sylink for directory ln -s %s %s \n", source, dest)
		err = os.Symlink(source, dest)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to create symlink from=%s to =%s", source, dest)
			os.Exit(1)
			return
		}
		return
	}

	// Move the file from the source to the directory
	err = moveOrCopyFile(source, dest)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error while moving/copying file to %s", dest)
		return
	}
	v.previousLevel = depth
}
