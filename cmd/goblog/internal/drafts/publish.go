package drafts

import (
	"context"
	"flag"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strconv"

	"github.com/fatih/color"
	"github.com/ldelossa/goblog"
)

var publishFS = flag.NewFlagSet("publish", flag.ExitOnError)

var publishFlags = struct{}{}

func publish(ctx context.Context) {
	publishFS.Usage = func() {
		fmt.Printf(`
The publish subcommand publishes an existing draft. 

When a draft is published it will be embedded into the next GoBlog binary created by running 'goblog publish'.

Usage:
	goblog drafts publish ID

`)
	}

	// 0: goblog, 1: drafts, 2: publish
	publishFS.Parse(os.Args[3:])

	if len(os.Args) < 4 {
		color.Red("Error: Not enough arguments provided to 'edit' subcommand\n")
		editFS.Usage()
		os.Exit(1)
	}

	// first arg must be id
	id, err := strconv.Atoi(os.Args[3])
	if err != nil {
		color.Red("Error: first argument to 'publish' subcommand must be an integer id")
		os.Exit(1)
	}

	sorted, err := sortedDrafts(ctx)
	if err != nil {
		color.Red("Error: failed retrieving drafts: %v", err)
		os.Exit(1)
	}
	if len(sorted) == 0 {
		color.Blue("There are no drafts to edit currently.\nUse 'goblog drafts new' to create one.")
		os.Exit(0)
	}

	if id == 0 {
		color.Red("Error: must provide a draft id.")
		os.Exit(1)
	}

	if id > len(sorted) {
		color.Red("Error: draft id %d does not exist", id)
		os.Exit(1)
	}

	draft := sorted[id-1]
	base := filepath.Base(draft.Path)

	err = os.Rename(
		path.Join(goblog.Drafts, base),
		path.Join(goblog.Posts, base),
	)
	if err != nil {
		color.Red(`
Error: failed to move draft to the posts directory: %v

`, err)
		os.Exit(1)
	}

}
