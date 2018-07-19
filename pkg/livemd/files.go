package livemd

import (
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/fsnotify/fsnotify"
)

var exts = []string{".markdown", ".md", ".mkd"} // has to be sorted in ascending order

// MarkdownFiles ...
func MarkdownFiles(searchDir string) ([]string, error) {
	fileList := []string{}
	err := filepath.Walk(searchDir, func(path string, f os.FileInfo, err error) error {
		ext := filepath.Ext(path)

		if i := sort.SearchStrings(exts, ext); i < len(exts) && strings.EqualFold(exts[i], ext) {
			absPath, err := filepath.Abs(path)
			if err != nil {
				return err
			}
			fileList = append(fileList, absPath)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return fileList, nil
}

// Watcher ...
func Watcher(files []string) (*fsnotify.Watcher, error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}
	for i := range files {
		watcher.Add(files[i])
	}
	return watcher, nil
}
