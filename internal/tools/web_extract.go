package tools

import (
	"net/url"
	"regexp"
	"strings"

	"golang.org/x/net/html"
)

var htmlWhitespace = regexp.MustCompile(`\s+`)

// extractHTML parses markup without running scripts. Markdown preserves the
// structure needed to read documentation, including links and code samples.
func extractHTML(source string, base *url.URL, markdown bool) (string, error) {
	doc, err := html.Parse(strings.NewReader(source))
	if err != nil {
		return "", err
	}
	root := doc
	var main *html.Node
	var find func(*html.Node)
	find = func(n *html.Node) {
		if main == nil && n.Type == html.ElementNode && (n.Data == "main" || n.Data == "article") {
			main = n
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			find(c)
		}
	}
	find(doc)
	if main != nil {
		root = main
	}
	var b strings.Builder
	var walk func(*html.Node)
	var plain func(*html.Node) string
	plain = func(n *html.Node) string {
		if n.Type == html.TextNode {
			return n.Data
		}
		var s strings.Builder
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			s.WriteString(plain(c))
		}
		return s.String()
	}
	children := func(n *html.Node) {
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			walk(c)
		}
	}
	attr := func(n *html.Node, key string) string {
		for _, a := range n.Attr {
			if a.Key == key {
				return a.Val
			}
		}
		return ""
	}
	walk = func(n *html.Node) {
		if n.Type == html.TextNode {
			b.WriteString(htmlWhitespace.ReplaceAllString(n.Data, " "))
			return
		}
		if n.Type != html.ElementNode {
			children(n)
			return
		}
		for _, a := range n.Attr {
			if a.Key == "hidden" || (a.Key == "aria-hidden" && a.Val == "true") {
				return
			}
		}
		switch n.Data {
		case "script", "style", "noscript", "template", "svg", "nav", "footer", "head":
			return
		}
		if n.Data == "pre" {
			text := plain(n)
			fence := "```"
			for strings.Contains(text, fence) {
				fence += "`"
			}
			b.WriteString("\n\n")
			if markdown {
				b.WriteString(fence + "\n")
			}
			b.WriteString(text)
			if markdown {
				b.WriteString("\n" + fence)
			}
			b.WriteString("\n\n")
			return
		}
		switch n.Data {
		case "h1", "h2", "h3", "h4", "h5", "h6":
			b.WriteString("\n\n")
			if markdown {
				b.WriteString(strings.Repeat("#", int(n.Data[1]-'0')) + " ")
			}
			children(n)
			b.WriteString("\n\n")
		case "p", "div", "section", "article", "main", "blockquote", "table":
			b.WriteString("\n\n")
			children(n)
			b.WriteString("\n\n")
		case "ul", "ol":
			b.WriteString("\n")
			children(n)
			b.WriteString("\n")
		case "li":
			b.WriteString("\n")
			if markdown {
				if n.Parent != nil && n.Parent.Data == "ol" {
					b.WriteString("1. ")
				} else {
					b.WriteString("- ")
				}
			}
			children(n)
		case "br", "hr":
			b.WriteString("\n")
		case "tr":
			b.WriteString("\n")
			children(n)
		case "td", "th":
			children(n)
			b.WriteString(" | ")
		case "strong", "b", "em", "i":
			marker := ""
			if markdown {
				marker = "*"
				if n.Data == "strong" || n.Data == "b" {
					marker = "**"
				}
			}
			b.WriteString(marker)
			children(n)
			b.WriteString(marker)
		case "code":
			if markdown {
				marker := "`"
				text := plain(n)
				for strings.Contains(text, marker) {
					marker += "`"
				}
				b.WriteString(marker + text + marker)
			} else {
				children(n)
			}
		case "a":
			label := strings.TrimSpace(plain(n))
			href, err := url.Parse(attr(n, "href"))
			if err == nil && base != nil {
				href = base.ResolveReference(href)
			}
			if markdown && err == nil && href.User == nil && (href.Scheme == "https" || href.Scheme == "http") && label != "" {
				b.WriteString("[" + strings.ReplaceAll(label, "]", "\\]") + "](<" + strings.ReplaceAll(href.String(), ">", "%3E") + ">)")
			} else {
				children(n)
			}
		case "img":
			b.WriteString(attr(n, "alt"))
		default:
			children(n)
		}
	}
	walk(root)
	return strings.TrimSpace(b.String()), nil
}
