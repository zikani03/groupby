package main

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
