package main

import "testing"

func TestNewDirectoryVisitor(t *testing.T) {
	tests := []struct {
		dir   string
		flat  bool
		depth int
	}{
		{dir: "TestRootDir", flat: true, depth: 5},
		{dir: "secondDir", flat: false, depth: 2},
	}

	for _, test := range tests {

		dv := NewDirectoryVisitor(test.dir, test.flat, test.depth)

		if dv.rootDir != test.dir {
			t.Errorf("NewDirectoryVisitor's rootDir is incorrect. Got '%s', Expected '%s'", dv.rootDir, test.dir)
		}
		if dv.flatten != test.flat {
			t.Errorf("NewDirectoryVisitor's flatten is incorrect. Got '%t', Expected '%t'", dv.flatten, test.flat)
		}
		if dv.maxDepth != test.depth {
			t.Errorf("NewDirectoryVisitor's maxDepth is incorrect. Got '%d', Expected '%d'", dv.maxDepth, test.depth)
		}
	}
}
