package main

import (
	"os"
	"testing"
	"time"
)

// fileInfo for testing the tree
type fileInfo struct {
	name    string
	size    int64
	mode    os.FileMode
	modTime time.Time
}

func (f fileInfo) Name() string {
	return f.name
}

func (f fileInfo) Size() int64 {
	return f.size
}

func (f fileInfo) Mode() os.FileMode {
	return f.mode
}

func (f fileInfo) ModTime() time.Time {
	return f.modTime
}

func (f fileInfo) IsDir() bool {
	return f.mode.IsDir()
}

func (f fileInfo) Sys() interface{} {
	return nil
}

func TestAddEntryIncrementsDirectoryAndFileCounts(t *testing.T) {
	pano := time.Now()
	fileMode := os.FileMode(0644)
	files := []fileInfo{
		fileInfo{"dir1", 1024, os.ModeDir, pano},
		fileInfo{"dir2", 1024, os.ModeDir, pano},
		fileInfo{"dir3", 1024, os.ModeDir, pano},
		fileInfo{"file1", 1024, fileMode, pano},
		fileInfo{"file2", 1024, fileMode, pano},
		fileInfo{"file3", 1024, fileMode, pano},
		fileInfo{"file4", 1024, fileMode, pano},
	}

	tree := &Tree{
		Root:           NewNode("/", pano.Year(), pano.Month(), pano.Day()),
		MaxDepth:       1,
		directoryCount: 0,
		fileCount:      0,
	}

	for _, f := range files {
		tree.AddEntry(f)
	}

	if tree.Directories() != 3 {
		t.Errorf("Tree's Directories() is incorrect. Got '%d', Expected '%d'", tree.Directories(), 3)
	}

	if tree.Files() != 4 {
		t.Errorf("Tree's Files() is incorrect. Got '%d', Expected '%d'", tree.Files(), 4)
	}
}
