package start

import (
	"os"
	"fmt"
	"path/filepath"
	"strings"
)

func walk(root string) []string {
	dirs := make([]string, 0)
	visit := func(path string, f os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		isHidden := strings.Contains(path, "/.")
		isVendor := strings.Contains(path, "/vendor/")
		if f.IsDir() && !isHidden && !isVendor {
			fmt.Printf("Visited: %s\n", path)
			dirs = append(dirs, path)
		}
		return nil
	}

	filepath.Walk(root, visit)
	return dirs
}
