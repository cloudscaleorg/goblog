package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"

	"github.com/fatih/color"
	"github.com/ldelossa/goblog"
	"github.com/ldelossa/goblog/cmd/goblog/internal/config"
	"github.com/ldelossa/goblog/cmd/goblog/internal/drafts"
	"github.com/ldelossa/goblog/cmd/goblog/internal/initialize"
	"github.com/ldelossa/goblog/cmd/goblog/internal/posts"
	"github.com/ldelossa/goblog/cmd/goblog/internal/serve"
)

const usage = `The goblog command line serves two purposes.
First it may act as an http server, serving assets and blog posts.
Secondly it helps you write and format blog posts. 

The command is split into subcommands, each containing their own help content.

goblog init    - create a new goblog environment
goblog config  - update configuration details
goblog serve   - serve your blog posts, assests, and web root over http
goblog posts   - list, view, and remove published posts
goblog drafts  - list, create, publish, and delete draft blog posts
goblog publish - build a new goblog binary with the latest posts and web root
goblog preview - preview your blog by running the code in $HOME/src directly
`

func main() {
	if len(os.Args) < 2 {
		fmt.Printf("Error: subcommand required\n\n")
		fmt.Println(usage)
		os.Exit(1)
	}

	switch os.Args[1] {
	case "--help":
		fmt.Println(usage)
		os.Exit(0)
	case "init":
		initialize.Initialize(context.TODO())
		os.Exit(0)
	case "config":
		config.Root(context.TODO())
	case "serve":
		serve.Serve()
	case "posts":
		posts.Root(context.TODO())
	case "drafts":
		drafts.Root(context.TODO())
	case "publish":
		_, err := initialize.NewBuildDecision().Exec(context.TODO())
		if err != nil {
			initialize.Initialize(context.TODO())
		}
	case "preview":
		cmd := exec.Command("go", "run", "./cmd/goblog/", "serve")
		cmd.Dir = goblog.Src
		cmd.Stdout = os.Stdout
		cmd.Stdin = os.Stdin
		cmd.Stderr = os.Stderr
		err := cmd.Run()
		if err != nil {
			color.Red("Error: failed to start goblog from src directory: %v", err)
			os.Exit(1)
		}
		os.Exit(0)
	default:
		fmt.Printf("Error: unrecognized subcommand: %s\n", os.Args[1])
		fmt.Println(usage)
	}
}
