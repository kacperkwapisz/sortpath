package fs

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
)

func Tree(dirPath string) (string, error) {
	var builder strings.Builder
	err := buildTree(&builder, dirPath, "")
	if err != nil {
		return "", err
	}
	return builder.String(), nil
}

func buildTree(builder *strings.Builder, dirPath, prefix string) error {
	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return err
	}
	// Sort entries: dirs first, then files, both alphabetically
	sort.Slice(entries, func(i, j int) bool {
		if entries[i].IsDir() == entries[j].IsDir() {
			return entries[i].Name() < entries[j].Name()
		}
		return entries[i].IsDir()
	})

	space := "    "
	branch := "│   "
	tee := "├── "
	last := "└── "

	for i, entry := range entries {
		pointer := tee
		if i == len(entries)-1 {
			pointer = last
		}
		builder.WriteString(prefix + pointer + entry.Name() + "\n")
		if entry.IsDir() {
			extension := branch
			if pointer == last {
				extension = space
			}
			nextPath := filepath.Join(dirPath, entry.Name())
			buildTree(builder, nextPath, prefix+extension)
		}
	}
	return nil
}
