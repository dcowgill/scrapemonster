package htmlutil

import (
	"code.google.com/p/cascadia"
	"code.google.com/p/go.net/html"
	"regexp"
	"strconv"
	"strings"
)

var (
	nonDigitsRegexp = regexp.MustCompile(`[^\d]+`)
)

// Tries to parse an integer from the text within a given node.
func ExtractInteger(root *html.Node, sel cascadia.Selector) *int {
	nodes := sel.MatchAll(root)
	if len(nodes) == 1 {
		s := RemoveNonDigits(TreeText(nodes[0]))
		if i, err := strconv.Atoi(s); err == nil {
			return &i
		}
	}
	return nil
}

// Removes all non-digit characters from a string.
func RemoveNonDigits(s string) string {
	return nonDigitsRegexp.ReplaceAllLiteralString(s, "")
}

// Returns the specified attribute of the node, or nil if not found.
func GetAttr(node *html.Node, key string) *html.Attribute {
	for _, a := range node.Attr {
		if a.Key == key {
			return &a
		}
	}
	return nil
}

// Returns the Data of the node's first immediate TextNode child.
func FirstText(n *html.Node) *string {
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if c.Type == html.TextNode {
			s := strings.TrimSpace(c.Data)
			return &s
		}
	}
	return nil
}

// Returns the joined Data of every TextNode found under the given node.
func TreeText(n *html.Node) string {
	var s string
	Visit(n, func(c *html.Node) {
		if c.Type == html.TextNode {
			s += c.Data
		}
	})
	return s
}

// Performs a post-order traversal of the given tree.
func Visit(tree *html.Node, f func(n *html.Node)) {
	for c := tree.FirstChild; c != nil; c = c.NextSibling {
		Visit(c, f)
	}
	f(tree)
}
