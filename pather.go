package wellington

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

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
