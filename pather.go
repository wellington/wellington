package wellington

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type stringSize []string

// Len returns length of stringSize
func (s stringSize) Len() int {
	return len(s)
}

// Less returns which string is shorter
func (s stringSize) Less(i, j int) bool {
	if len(s[i]) < len(s[j]) {
		return true
	}
	return false
}

// Swap replaces the two items indicated by indexes
func (s stringSize) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

// relative takes the input paths and file one that file
// can be made relative to. This creates a relative path useful for
// writing out build files
//
// It is expected that a file is part of one of the path directories.
// Otherwise, it will not match any of them.
func relative(paths []string, file string) string {

	if len(filepath.Ext(file)) > 0 {
		file = filepath.Dir(file)
	}

	for _, path := range paths {
		if len(filepath.Ext(path)) > 0 {
			path = filepath.Dir(path)
		}

		if !strings.HasPrefix(file, path) {
			continue
		}
		rel, err := filepath.Rel(path, file)

		if err == nil {
			return rel
		}
	}
	return ""
}

func pathsToFiles(paths []string, recurse bool) []string {
	var rollup []string
	for _, path := range paths {
		if recurse {
			paths, err := recursePath(path)
			if err != nil {
				fmt.Println(err)
				continue
			}
			rollup = append(rollup, paths...)
		} else {
			rollup = append(rollup, resolvePath(path)...)
		}
	}
	return rollup
}

var resolveExts = []string{".scss", ".sass"}

func isImportable(name string) bool {
	ext := filepath.Ext(name)

	var match bool
	for _, r := range resolveExts {
		if ext == r {
			match = true
		}
	}

	// remove partials
	return match && !strings.HasPrefix(name, "_")
}

// recursePath takes an input path and locates all non-partial scss/sass
// files in it
func recursePath(path string) ([]string, error) {
	var paths []string
	err := filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
		if info == nil {
			return fmt.Errorf("invalid file found: %s", path)
		}
		if !info.IsDir() && isImportable(info.Name()) {
			paths = append(paths, path)
		}
		return nil
	})
	return paths, err
}

//
func resolvePath(path string) []string {
	var paths []string
	for _, r := range resolveExts {
		some, err := filepath.Glob(filepath.Join(path, "*"+r))
		if err != nil {
			continue
		}
		paths = append(paths, some...)
	}

	return paths
}
