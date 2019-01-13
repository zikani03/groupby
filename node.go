package main

import "time"

type Node struct {
	FileName string
	Year     int
	Month    time.Month
	Day      int
	Next     *Node
	Children *Node
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
