package ast

import (
	"regexp"
)

var aliasRegex = regexp.MustCompile(`\(alias ([^)]+)\)`)

func aliasFromMeta(meta string) string {
	match := aliasRegex.FindStringSubmatch(meta)

	if match != nil {
		return match[1]
	} else {
		return ""
	}
}

type lineHandler struct {
	Matcher       *regexp.Regexp
	MatchCallback func(matches []string, node *AstNode)
}

var typeOnlyHandler = lineHandler{
	Matcher:       regexp.MustCompile("^( *)([^ ]+)$"),
	MatchCallback: func(matches []string, node *AstNode) {},
}

var typeWithValueHandler = lineHandler{
	Matcher: regexp.MustCompile("^( *)([^ ]+) ([^ ]*)$"),
	MatchCallback: func(matches []string, node *AstNode) {
		node.Value = matches[3]
	},
}

var typeWithMetaHandler = lineHandler{
	Matcher: regexp.MustCompile(`^( *)([^ ]+) \((.+)\)$`),
	MatchCallback: func(matches []string, node *AstNode) {
		node.Meta = matches[3]
		node.Alias = aliasFromMeta(matches[3])
	},
}

var typeWithValueAndMetaHandler = lineHandler{
	Matcher: regexp.MustCompile(`^( *)([^ ]+) ([^ ]*) (\(.+\))$`),
	MatchCallback: func(matches []string, node *AstNode) {
		node.Value = matches[3]
		node.Meta = matches[4]
		node.Alias = aliasFromMeta(matches[4])
	},
}

var allHandlers = []lineHandler{
	typeOnlyHandler,
	typeWithValueHandler,
	typeWithMetaHandler,
	typeWithValueAndMetaHandler,
}

// TODO: simplify handleMatch
func applyLineHandlers(line string, handleMatch func(r *regexp.Regexp, line string, cb func(matches []string, line *AstNode)) bool) bool {
	for _, handler := range allHandlers {
		if handleMatch(handler.Matcher, line, handler.MatchCallback) {
			return true
		}
	}

	return false
}

// ultimate goal will be to construct a tree of AST nodes
// with root nodes also containing the query string.
// the roots of the tree represent the top-level queries
func Parse(lines []string) (root *AstNode, err error) {
	var previousLine *AstNode

	handleMatch := func(r *regexp.Regexp, line string, cb func(matches []string, line *AstNode)) bool {
		matches := r.FindStringSubmatch(line)

		if matches != nil {
			parsedLine := AstNode{RawLine: line}
			parsedLine.Indent = len(matches[1])
			parsedLine.Type = matches[2]
			var previousParent *AstNode

			if previousLine != nil {
				previousParent = previousLine.Parent
			}

			cb(matches, &parsedLine)

			if previousLine != nil {
				if previousLine.Indent < parsedLine.Indent {
					previousLine.Children = append(previousLine.Children, &parsedLine)
					parsedLine.Parent = previousLine
				} else if previousLine.Indent == parsedLine.Indent && previousParent != nil {
					previousParent.Children = append(previousParent.Children, &parsedLine)
					parsedLine.Parent = previousParent
				} else if previousLine.Indent > parsedLine.Indent {
					var previousParentParent *AstNode

					if previousParent.Parent != nil {
						previousParentParent = previousParent.Parent

						for previousParentParent.Indent != parsedLine.Indent-1 {
							previousParentParent = previousParentParent.Parent
							if previousParentParent == nil {
								panic("Could not find parent parent")
							}
						}

						previousParentParent.Children = append(previousParentParent.Children, &parsedLine)
						parsedLine.Parent = previousParentParent
					}
				}
			} else {
				root = &parsedLine
			}

			previousLine = &parsedLine

			return true
		}

		return false
	}

	for _, line := range lines {
		if line == "" {
			continue
		}

		anyMatch := applyLineHandlers(line, handleMatch)

		if !anyMatch {
			panic("No match for line: " + line)
		}
	}

	return
}
