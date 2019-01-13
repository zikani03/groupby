package main

type NodeVisitor interface {
	Visit(n *Node, depth int)
}
