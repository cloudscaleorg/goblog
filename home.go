package goblog

import (
	"os"
	"os/user"
	"path"

	"github.com/fatih/color"
)

var (
	// Home is the root folder GoBlog works out of.
	// It provides a well defined home enviroment
	// where other directories can root themselves.
	//
	// Resolving a desired Home directory is dependent
	// on user.Current() returning the current user.
	Home string
	// Src is a directory nested in Home where GoBlog's
	// downstream (forked) source code lives.
	Src string
	// Posts is a directory which hold published
	// GoBlog posts
	Posts string
	// Drafts is a directory which holds draft blog
	// posts until they are published.
	Drafts string
)

func init() {
	// we need to be able to determine the current user
	// to resolve home dirs.
	usr, err := user.Current()
	if err != nil {
		color.Red("Error: GoBlog must be able to determine the current user.")
		os.Exit(1)
	}
	Home = path.Join(usr.HomeDir, "goblog")
	Src = path.Join(Home, "src")
	Posts = path.Join(Src, "posts")
	Drafts = path.Join(Home, "src", "drafts")
}
