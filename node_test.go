package main

import (
	"testing"
	"time"
)

type nodeDetails struct {
	file        string
	year        int
	month       time.Month
	day         int
	hasNext     bool
	hasChildren bool
}

// Tests NewNode, Node.HasNext, Node.HasChildren, Node.AddChild
func TestNode(t *testing.T) {

	// Create parent node
	n := NewNode("/root", 2010, time.April, 9)
	assertNode(t, n, nodeDetails{
		file: "/root", year: 2010, month: time.April, day: 9, hasNext: false, hasChildren: false,
	})

	// Create child node
	n.AddChild(NewNode("/child", 2018, time.March, 15))

	assertNode(t, n, nodeDetails{
		file: "/root", year: 2010, month: time.April, day: 9, hasNext: false, hasChildren: true,
	})
	assertNode(t, n.Children, nodeDetails{
		file: "/child", year: 2018, month: time.March, day: 15, hasNext: false, hasChildren: false,
	})

	// Create second child node
	n.AddChild(NewNode("/child2", 2016, time.November, 30))
	assertNode(t, n, nodeDetails{
		file: "/root", year: 2010, month: time.April, day: 9, hasNext: false, hasChildren: true,
	})
	assertNode(t, n.Children, nodeDetails{
		file: "/child2", year: 2016, month: time.November, day: 30, hasNext: true, hasChildren: false,
	})
	assertNode(t, n.Children.Next, nodeDetails{
		file: "/child", year: 2018, month: time.March, day: 15, hasNext: false, hasChildren: false,
	})
}

func assertNode(t *testing.T, n *Node, expects nodeDetails) {
	// Node Constructor
	if n.FileName != expects.file ||
		n.Year != expects.year ||
		n.Month != expects.month ||
		n.Day != expects.day {
		t.Errorf("Node's properties do not match expected. Expected (%s, %d %d %d), Got (%s, %d %d %d)",
			expects.file, expects.year, expects.month, expects.day,
			n.FileName, n.Year, n.Month, n.Day)
	}

	if expects.hasNext != n.HasNext() {
		if expects.hasNext {
			t.Errorf("Node '%s' does not have expected Next node.", n.FileName)
		} else {
			t.Errorf("Node '%s' has expected Next node.", n.FileName)
		}
	}

	if expects.hasChildren != n.HasChildren() {
		if expects.hasChildren {
			t.Errorf("Node '%s' does not have expected Children node.", n.FileName)
		} else {
			t.Errorf("Node '%s' has expected Children node.", n.FileName)
		}
	}
	if n.HasChildren() != expects.hasChildren {
		t.Errorf("Node '%s' has children when it should not.", n.FileName)
	}
}

// Tests Node.Search
func TestNodeSearch(t *testing.T) {
	n := NewNode("/root", 2010, time.April, 9)
	n.AddChild(NewNode("/child", 2018, time.March, 15))
	n.AddChild(NewNode("/child2", 2016, time.November, 30))

	tests := []struct {
		input           string
		expectedFound   bool
		expectedDetails nodeDetails
	}{
		{input: "nonexistantNode", expectedFound: false, expectedDetails: nodeDetails{}},
		{input: "/root", expectedFound: false, expectedDetails: nodeDetails{}},
		{input: "/child", expectedFound: false, expectedDetails: nodeDetails{
			file: "/child", year: 2018, month: time.March, day: 15, hasNext: false, hasChildren: false,
		}},
		{input: "/child2", expectedFound: false, expectedDetails: nodeDetails{
			file: "/child2", year: 2016, month: time.November, day: 30, hasNext: true, hasChildren: false,
		}},
	}

	for _, test := range tests {
		res := n.Search(test.input)
		if res == nil {
			if test.expectedFound {
				t.Errorf("Node '%s' expected, but not found.", test.input)
			}
			continue
		}

		assertNode(t, res, test.expectedDetails)
	}
}
