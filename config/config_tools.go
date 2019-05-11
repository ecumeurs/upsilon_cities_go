package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

var root string
var rootSlash string

//Root seek root path of the project.
func Root() string {
	if root == "" {
		if SYS_FORCE_ROOT {
			return SYS_ROOT
		}

		// get working directory and go up until finding upsilon_cities_go
		dir, _ := os.Getwd()
		dir = filepath.ToSlash(dir)
		subs := strings.Split(dir, "/")

		// keep what's necessary up until last inserted is upsilon_cities_go
		nsubs := make([]string, 0)

		for _, v := range subs {
			nsubs = append(nsubs, v)
			if v == "upsilon_cities_go" {
				break
			}

		}

		rootSlash = strings.Join(nsubs, "/")
		root = filepath.FromSlash(rootSlash)
	}
	return root
}

//MakePath create a path from root to target.
func MakePath(to string) string {
	Root()
	return filepath.FromSlash(fmt.Sprintf("%s/%s", rootSlash, to))
}
