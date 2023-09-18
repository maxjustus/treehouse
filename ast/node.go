package ast

// TODO: add hash
type AstNode struct {
	RawLine string
	Indent  int
	Type    string
	Value   string
	// Think table name, database name, etc.
	ValueQualifier string
	Alias          string
	Meta           string
	// Hash int64 // hash of the node and
	// children to quickly compare branches
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

func (node *AstNode) descendentOf(parentType string) bool {
	return node.nearestParentOfType(parentType) != nil
}

func (node *AstNode) nearestParentOfType(parentType string) *AstNode {
	if node.Parent == nil {
		return nil
	} else if node.Parent.Type == parentType {
		return node.Parent
	} else {
		return node.Parent.nearestParentOfType(parentType)
	}
}

func (node *AstNode) nearestSelectTableNodes() []*AstNode {
	nearestSelect := node.nearestParentOfType("SelectQuery")
	if nearestSelect == nil {
		return nil
	}

	return nearestSelect.childrenOfType("TableIdentifier")
}

func (node *AstNode) childrenOfType(nodeType string) []*AstNode {
	var children []*AstNode

	for _, child := range node.Children {
		if child.Type == nodeType {
			children = append(children, child)
		}
	}

	return children
}

func (node *AstNode) parentIsType(parentType string) bool {
	return node.Parent != nil && node.Parent.Type == parentType
}

func (node *AstNode) parentIsAlterColumn(commandType string) bool {
	return node.parentIsType("AlterCommand") && node.Parent.Value == commandType
}
