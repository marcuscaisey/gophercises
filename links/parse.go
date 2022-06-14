package links

import (
	"fmt"
	"io"
	"strings"

	"golang.org/x/net/html"
)

type Link struct {
	Href string
	Text string
}

func (l Link) String() string {
	return fmt.Sprintf(`<a href="%s">%s</a>`, l.Href, l.Text)
}

func Parse(r io.Reader) ([]Link, error) {
	doc, err := html.Parse(r)
	if err != nil {
		return nil, fmt.Errorf("parse HTML: %s", err)
	}
	return parse(doc), nil
}

func parse(root *html.Node) []Link {
	if root.Data == "a" {
		return []Link{newLinkFromAnchor(root)}
	}

	var links []Link
	for child := root.FirstChild; child != nil; child = child.NextSibling {
		links = append(links, parse(child)...)
	}
	return links
}

func newLinkFromAnchor(node *html.Node) Link {
	var href string
	for _, attr := range node.Attr {
		if attr.Key == "href" {
			href = attr.Val
			break
		}
	}
	return Link{
		Href: href,
		Text: nodeText(node),
	}
}

func nodeText(root *html.Node) string {
	if root.Type == html.TextNode {
		return strings.TrimSpace(root.Data)
	}

	var parts []string
	for child := root.FirstChild; child != nil; child = child.NextSibling {
		parts = append(parts, nodeText(child))
	}
	return strings.Join(parts, "")
}
