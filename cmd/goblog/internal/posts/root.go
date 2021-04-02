package posts

import (
	"context"
	"fmt"
	"os"

	"github.com/fatih/color"
	"github.com/ldelossa/goblog/cmd/goblog/internal/initialize"
)

var usage = `The 'posts' subcommand is for managing published blog posts.
These posts are embedded into the GoBlog binary.
If you're removing a post you'll need to rebuild GoBlog.

The '--local' flag optionally instructs GoBlog is look at local posts, ones not
embedded into the binary.

Usage: 

goblog posts [--local] subcommand

goblog posts list  - list published blog posts and their id
goblog posts view  - view the markdown contents of a post
goblog posts draft - unpublish a post and move it to draft (assumes --local flag)
`

// Root is the 'posts' subcommand root handler
func Root(ctx context.Context) {
	if len(os.Args) < 3 {
		color.Red(`
Error: The 'posts' subcommand requires a directive.

`)
		color.Blue(usage)
		os.Exit(1)
	}

	var local bool

	if os.Args[2] == "--local" || os.Args[2] == "-local" {
		if len(os.Args) < 4 {
			color.Red(`
Error: The 'posts' subcommand requires a directive.

`)
			color.Blue(usage)
			os.Exit(1)
		}
		local = true
		// pop off arg
		os.Args = append(os.Args[:2], os.Args[3:]...)
	}

	switch os.Args[2] {
	case "--help":
		color.Blue(usage)
		os.Exit(0)
	case "list":
		// if we are listing local posts, we need to ensure
		// GoBlog's home is initialized.
		if local {
			initialize.Initialize(context.TODO())
		}
		list(ctx, local)
	case "draft":
		// If we are moving a post from published to draft,
		// we need to ensure GoBlog's home is initialized.
		initialize.Initialize(context.TODO())
		draft(ctx, local)
	case "view":
		// if we are viewing local posts, we need to ensure
		// GoBlog's home is initialized.
		if local {
			initialize.Initialize(context.TODO())
		}
		view(ctx, local)
	default:
		fmt.Printf(`
Error: unrecognized subcommand: %s

`, os.Args[2])
		os.Exit(1)
	}
}
