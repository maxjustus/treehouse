package ast

type AstNode struct {
	RawLine  string
	Indent   int
	Type     string
	Value    string
	Alias    string
	Meta     string
	Parent   *AstNode
	Children []*AstNode
}

func (n *AstNode) Walk(cb func(node *AstNode)) {
	cb(n)
	if n.Children != nil {
		for _, child := range n.Children {
			child.Walk(cb)
		}
	}
}
