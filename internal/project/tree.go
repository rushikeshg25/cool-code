package project

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type treeNode struct {
	name     string
	isDir    bool
	children []treeNode
}

// FolderStructure renders an ASCII tree of rootDir, skipping gitignored paths.
// maxDepth < 0 means unlimited.
func FolderStructure(rootDir string, checker GitIgnoreChecker, maxDepth int) string {
	if checker == nil {
		checker = func(string) bool { return false }
	}
	nodes := buildTree(rootDir, rootDir, checker, 0, maxDepth)
	var b strings.Builder
	b.WriteString(filepath.Base(rootDir) + "/\n")
	formatTree(&b, nodes, "")
	return b.String()
}

func buildTree(dir, base string, checker GitIgnoreChecker, depth, maxDepth int) []treeNode {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil
	}
	sort.Slice(entries, func(i, j int) bool {
		di, dj := entries[i].IsDir(), entries[j].IsDir()
		if di != dj {
			return di
		}
		return entries[i].Name() < entries[j].Name()
	})

	var nodes []treeNode
	for _, entry := range entries {
		full := filepath.Join(dir, entry.Name())
		rel, _ := filepath.Rel(base, full)
		if checker(rel) {
			continue
		}
		node := treeNode{name: entry.Name(), isDir: entry.IsDir()}
		if entry.IsDir() && (maxDepth < 0 || depth < maxDepth) {
			node.children = buildTree(full, base, checker, depth+1, maxDepth)
		}
		nodes = append(nodes, node)
	}
	return nodes
}

func formatTree(b *strings.Builder, nodes []treeNode, prefix string) {
	for i, node := range nodes {
		last := i == len(nodes)-1
		branch := "├───"
		nextPrefix := prefix + "│   "
		if last {
			branch = "└───"
			nextPrefix = prefix + "    "
		}
		suffix := ""
		if node.isDir {
			suffix = "/"
		}
		b.WriteString(prefix + branch + node.name + suffix + "\n")
		if node.isDir && len(node.children) > 0 {
			formatTree(b, node.children, nextPrefix)
		}
	}
}
