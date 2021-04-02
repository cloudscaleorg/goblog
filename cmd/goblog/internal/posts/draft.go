package posts

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

var draftFS = flag.NewFlagSet("draft", flag.ExitOnError)

var draftFlags = struct {
}{}

func draft(ctx context.Context, local bool) {
	listFS.Usage = func() {
		fmt.Printf(`
The 'draft' subcommand moves published posts into drafts.

If the '--local' flag is used it will target only local posts not embedded in a GoBlog binary yet.

Unless the '--local' flag was used, you will need to issue a 'goblog publish' to see the results in a new GoBlog binary.

Usage:
	goblog posts draft ID
`)
	}

	// 0: goblog 1: posts 2: draft
	if len(os.Args) < 4 {
		color.Red(`
Error: Not enough arguments to 'draft' subcommand

`)
		listFS.Usage()
		os.Exit(1)
	}

	id, err := strconv.Atoi(os.Args[3])
	if err != nil {
		color.Red(`
Error: 'draft' subcommand requires an integer ID argument

`)
		listFS.Usage()
		os.Exit(1)
	}

	var posts goblog.DateSortable
	if local {
		posts, err = sortedLocalPosts(ctx)
		if err != nil {
			color.Red(`
Error: failed to get local posts: %v

`, err)
			os.Exit(1)
		}
	} else {
		posts = goblog.DSCache
	}

	if len(posts) == 0 {
		color.Blue(`
There are no posts to view currently.

Use 'goblog drafts new' to create one and 'goblog build' to build a GoBlog binary with your new posts.

`)
		os.Exit(0)
	}

	if id == 0 {
		color.Red(`
Error: must provide a post id.

`)
		os.Exit(1)
	}

	if id > len(posts) {
		color.Red(`
id not found

`)
		os.Exit(1)
	}

	post := posts[id-1]
	base := filepath.Base(post.Path)
	err = os.Rename(path.Join(goblog.Posts, base), path.Join(goblog.Drafts, base))
	if err != nil {
		color.Red(`
Error: failed to move post from posts dir to drafts dir: %v

`, err)
		os.Exit(1)
	}

	color.Blue(`
Post %d successfully moved to drafts.

Use 'goblog publish' to publish a GoBlog binary with this post removed.
`)
	os.Exit(0)
}
