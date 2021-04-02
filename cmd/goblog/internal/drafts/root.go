package drafts

import (
	"context"
	"fmt"
	"os"

	"github.com/fatih/color"
	"github.com/ldelossa/goblog/cmd/goblog/internal/initialize"
)

var usage = `The 'drafts' subcommand is used to edit and publish local drafts.

Drafts are your working space where you can leave ideas hanging around. 

Once a draft is finished it is 'published' and will be embedded into the next GoBlog binary you build via the 'goblog publish' command.

If you are looking to work with published posts see the 'goblog posts' command instead.

goblog drafts new     - create and edit a new draft
goblog drafts edit    - edit an existing draft or its metadata
goblog drafts list    - list drafts 
goblog drafts view    - view the contents of a draft
goblog drafts delete  - delete a draft
goblog drafts publish - publishes a draft 
`

// Root is the 'drafts' subcommand root handler.
func Root(ctx context.Context) {
	if len(os.Args) < 3 {
		color.Red("Error: The 'drafts' subcommand requires a directive.")
		color.Blue(usage)
		os.Exit(1)
	}
	if os.Args[2] == "--help" || os.Args[2] == "-help" {
		fmt.Printf(usage)
		os.Exit(0)
	}
	initialize.Initialize(context.TODO())
	switch os.Args[2] {
	case "edit":
		edit(ctx)
	case "new":
		new(ctx)
	case "list":
		list(ctx)
	case "view":
		view(ctx)
	case "delete":
		delete(ctx)
	case "publish":
		publish(ctx)
	default:
		color.Red(`
Error: unknown subcommand provided.

`)
		fmt.Printf(usage)
		os.Exit(1)
	}
}
