package drafts

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/ldelossa/goblog"
	"gopkg.in/yaml.v3"
)

var editFS = flag.NewFlagSet("edit", flag.ExitOnError)

var editFlags = struct {
}{}

func edit(ctx context.Context) {
	editFS.Usage = func() {
		fmt.Printf(`
The edit subcommand opens a draft for editing. 

The '--meta' flag may be used to edit both the contents and metadata of a post in yaml syntax.

Usage:
	goblog drafts edit ID [--meta]

`)
	}
	// 0: goblog, 1: drafts, 2: edit
	editFS.Parse(os.Args[3:])

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

	var meta bool
	for _, arg := range os.Args {
		if arg == "--meta" || arg == "-meta" {
			meta = true
		}
	}

	scanner := bufio.NewScanner(os.Stdin)
	editor := os.Getenv("EDITOR")
	if editor == "" {
		color.Red("Please set your EDITOR environment variable.")
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

	// handle meta edit only
	if meta {
		// call editor
		var cmd *exec.Cmd
		cmd = exec.Command(editor, draft.Path)
		cmd.Stdout = os.Stdout
		cmd.Stdin = os.Stdin
		err = cmd.Run()
		if err != nil {
			color.Red("Error: failed to start editor: %v", err)
			os.Exit(1)
		}
		os.Exit(0)
	}

	mdDraft := strings.ReplaceAll(draft.Path, ".post", ".md")

	f, err := os.OpenFile(mdDraft, os.O_CREATE|os.O_RDWR, 0o0660)
	if err != nil {
		color.Red("Error: failed opening scratch markdown file: %v", err)
		os.Exit(1)
	}

	_, err = io.WriteString(f, draft.MarkDown.Value)
	if err != nil {
		color.Red("Error: failed copying draft to scratch markdown file: %v", err)
		os.Exit(1)
	}
	err = f.Close()
	if err != nil {
		color.Red("Error: failed closing scratch markdown file: %v", err)
		os.Exit(1)
	}

	// call editor
	var cmd *exec.Cmd
	cmd = exec.Command(editor, mdDraft)
	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin
	err = cmd.Run()
	if err != nil {
		color.Red("Error: failed to start editor: %v", err)
		os.Exit(1)
	}

	// read markdown file into a buffer, close fd and remove
	// the draft.
	f, err = os.OpenFile(mdDraft, os.O_CREATE|os.O_RDWR, 0o0660)
	if err != nil {
		color.Red("Error: failed to open scratch markdown file: %v", err)
		os.Exit(1)
	}
	buff, err := io.ReadAll(f)
	if err != nil {
		color.Red("Error: failed reading your markdown draft: %v", err)
		os.Exit(1)
	}
	err = f.Close()
	if err != nil {
		color.Red("Error: failed closing your markdown draft: %v", err)
		os.Exit(1)
	}
	err = os.Remove(mdDraft)
	if err != nil {
		color.Red("Error: failed reading file back: %v", err)
	}

	draft.MarkDown = yaml.Node{
		Kind:  yaml.ScalarNode,
		Style: yaml.FlowStyle,
		Value: string(buff),
	}
	draft.Date = time.Now()

	// ask user if they want to publish this or keep draft
	var publish string
	color.Yellow("Publish this post? ('yes' or 'no')")
	fmt.Printf("> ")
	scanner.Scan()
	if err := scanner.Err(); err != nil {
		color.Red("Error: failed scanning input: %v", err)
		color.Red("Dumping your markdown so you don't loose your work...\n")
		fmt.Println("----MARKDOWN BEGIN----")
		fmt.Printf(string(buff))
		fmt.Println("----MARKDOWN END----")
		os.Exit(1)
	}
	publish = scanner.Text()

	// draft.Path will already be formated and have
	// .post syntax
	formated := path.Base(draft.Path)
	var postPath string
	switch publish {
	case "yes":
		postPath = path.Join(goblog.Posts, formated)
	default:
		postPath = path.Join(goblog.Drafts, formated)
	}

	f, err = os.OpenFile(postPath, os.O_CREATE|os.O_WRONLY, 0o0660)
	if err != nil {
		color.Red("Error: failed to create GoBlog post file: %v", err)
		color.Red("Dumping your markdown so you don't loose your work...")
		fmt.Println("----MARKDOWN BEGIN----")
		fmt.Printf(string(buff))
		fmt.Println("----MARKDOWN END----")
		os.Exit(1)
	}

	err = yaml.NewEncoder(f).Encode(draft)
	if err != nil {
		color.Red("Error: failed to create GoBlog post file: %v", err)
		color.Red("Dumping your markdown so you don't loose your work...")
		fmt.Println("----MARKDOWN BEGIN----")
		fmt.Printf(string(buff))
		fmt.Println("----MARKDOWN END----")
		os.Exit(1)
	}

	color.Blue(`
Your draft has been written to: %v. 

`, postPath)

	// if we published this draft, delete it from drafts folder
	if publish == "yes" {
		err = os.Remove(path.Join(goblog.Drafts, formated))
		if err != nil {
			color.Red(`
Error: failed to remove draft: %v

Your post was still published but will remain viewable in the draft list.

`, err)
		}

	}
	os.Exit(0)
}
