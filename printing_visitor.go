package main

import "fmt"

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

	prefix := SubdirectoryInner
	if !n.HasNext() {
		prefix = SubdirectoryLink
	}

	filename := FileNameByDepth(n.FileName, depth)

	fmt.Println(prefix, filename)

	p.previousLevel = depth
}
