package posts

import (
	"context"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strconv"

	"github.com/fatih/color"
	"github.com/ldelossa/goblog"
	"gopkg.in/yaml.v3"
)

var viewFS = flag.NewFlagSet("view", flag.ExitOnError)

var viewFlags = struct {
}{}

func view(ctx context.Context, local bool) {
	viewFS.Usage = func() {
		fmt.Printf(`
The view subcommand prints post contents to stdout. 

The '--meta' flag may be used to print both the post content and the metdata data in yaml syntax.

Usage:
	goblog posts view ID [--meta]

`)
	}

	if len(os.Args) < 4 {
		color.Red("Error: Not enough arguments to 'view' subcommand\n")
		listFS.Usage()
		os.Exit(1)
	}

	// first arg must be id
	id, err := strconv.Atoi(os.Args[3])
	if err != nil {
		color.Red("Error: first argument to 'view' subcommand must be an integer id")
		os.Exit(1)
	}

	var meta bool
	for _, arg := range os.Args {
		if arg == "--meta" || arg == "-meta" {
			meta = true
		}
	}

	var posts goblog.DateSortable
	if local {
		posts, err = sortedLocalPosts(ctx)
		if err != nil {
			color.Red("Error: failed to get local posts: %v", err)
			os.Exit(1)
		}
	} else {
		posts = goblog.DSCache
	}

	if len(posts) == 0 {
		color.Blue(`
There are no posts to view currently.

Use 'goblog drafts new' to create one and 'goblog publish' to build a GoBlog binary with your new posts.

`)
		os.Exit(0)
	}

	if id == 0 {
		color.Red("Error: must provide a post id.")
		os.Exit(1)
	}

	if id > len(posts) {
		fmt.Println("id not found")
		os.Exit(1)
	}

	post := posts[id-1]

	var f fs.File
	if local {
		f, err = os.Open(post.Path)
		if err != nil {
			fmt.Println("error viewing post: " + err.Error())
			os.Exit(1)
		}
	} else {
		f, err = goblog.PostsFS.Open(post.Path)
		if err != nil {
			fmt.Println("error viewing post: " + err.Error())
			os.Exit(1)
		}
	}
	// just write out the file data and exit 0
	if meta {
		_, err := io.Copy(os.Stdout, f)
		if err != nil {
			fmt.Println("error viewing post: " + err.Error())
			os.Exit(1)
		}
		os.Exit(0)
	}

	err = yaml.NewDecoder(f).Decode(&post)
	if err != nil {
		fmt.Println("error viewing post: " + err.Error())
		os.Exit(1)
	}

	if meta {
		fmt.Println(post.MarkDown.Value)
	}
	fmt.Println(post.MarkDown.Value)
}

// walks the "posts" directory for local posts, sorts them by date, and returns
// a list of them.
func sortedLocalPosts(ctx context.Context) (goblog.DateSortable, error) {
	var sorted goblog.DateSortable
	err := filepath.Walk(goblog.Posts, func(path string, info os.FileInfo, err error) error {
		if info.IsDir() {
			return nil
		}

		f, err := os.Open(path)
		if err != nil {
			return err
		}

		var post goblog.Post
		err = yaml.NewDecoder(f).Decode(&post)
		if err != nil {
			return err
		}
		// ignore the empty post
		if post.Title == "_empty" {
			return nil
		}
		post.Path = path
		sorted = append(sorted, post)
		return nil
	})
	if err != nil {
		return sorted, err
	}
	return sorted, nil
}
