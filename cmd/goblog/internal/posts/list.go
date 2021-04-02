package posts

import (
	"context"
	"flag"
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/fatih/color"
	"github.com/ldelossa/goblog"
)

var listFS = flag.NewFlagSet("list", flag.ExitOnError)

var listFlags = struct {
}{}

func list(ctx context.Context, local bool) {
	listFS.Usage = func() {
		fmt.Printf(`
The 'list' subcommand will list posts in date order.

If the '--local' flag is used a list of local posts, ones not emedded into the binary, will be listed.

This subcommand takes no arguments.
`)
	}

	// 0: goblog, 1: posts 3: list
	listFS.Parse(os.Args[3:])

	var posts goblog.DateSortable
	var err error
	if local {
		posts, err = sortedLocalPosts(ctx)
		if err != nil {
			color.Red("Error: failed to query local posts: %v", err)
			os.Exit(1)
		}
	} else {
		posts = goblog.DSCache
	}

	if len(posts) == 0 {
		fmt.Println("No posts found.")
	}

	tw := tabwriter.NewWriter(os.Stdout, 0, 8, 1, '\t', 0)
	fmt.Fprintln(tw, "ID\tDATE\tTITLE\tSUMMARY")
	for i, post := range posts {
		fmt.Fprintf(tw, "%d\t%s\t%s\t%s\n", i+1, post.Date.Format("2006-Jan-2"), post.Title, post.Summary)
	}
	err = tw.Flush()
	if err != nil {
		fmt.Println("error: " + err.Error())
		os.Exit(1)
	}
	return
}
