package drafts

import (
	"context"
	"flag"
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/fatih/color"
)

var listFS = flag.NewFlagSet("list", flag.ExitOnError)

var listFlags = struct {
}{}

func list(ctx context.Context) {
	listFS.Usage = func() {
		fmt.Printf(`
The list subcommand lists drafts and their ids. 

This subcommand takes no arguments.

Usage:
	goblog drafts list
`)
	}
	sorted, err := sortedDrafts(ctx)
	if err != nil {
		color.Red("")
		os.Exit(1)
	}
	tw := tabwriter.NewWriter(os.Stdout, 0, 8, 1, '\t', 0)
	fmt.Fprintln(tw, "ID\tDATE\tTITLE\tSUMMARY")
	for i, draft := range sorted {
		fmt.Fprintf(tw, "%d\t%s\t%s\t%s\n", i+1, draft.Date.Format("2006-Jan-2"), draft.Title, draft.Summary)
	}
	err = tw.Flush()
	if err != nil {
		fmt.Println("error: " + err.Error())
		os.Exit(1)
	}
	return
}
