package drafts

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strconv"

	"github.com/fatih/color"
)

var deleteFS = flag.NewFlagSet("delete", flag.ExitOnError)

var deleteFlags = struct {
}{}

func delete(ctx context.Context) {
	deleteFS.Usage = func() {
		fmt.Printf(`
The delete subcommand removes a draft.

Usage:
	goblog drafts delete ID
`)
	}

	if len(os.Args) < 4 {
		color.Red("Error: Not enough arguments provided to 'edit' subcommand\n")
		editFS.Usage()
		os.Exit(1)
	}

	// first arg must be id
	id, err := strconv.Atoi(os.Args[3])
	if err != nil {
		color.Red("Error: first argument to 'view' subcommand must be an integer id")
		os.Exit(1)
	}

	sorted, err := sortedDrafts(ctx)
	if err != nil {
		color.Red("Error: failed retrieving drafts: %v", err)
		os.Exit(1)
	}
	if len(sorted) == 0 {
		color.Blue(`
There are no drafts to edit currently.

Use 'goblog drafts new' to create one.

`)
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

	err = os.Remove(draft.Path)
	if err != nil {
		color.Red("Error: failed to remove your draft: %v", err)
		os.Exit(1)
	}
	color.Blue(`
Successfully deleted draft %v

`, id)
	os.Exit(0)
}
