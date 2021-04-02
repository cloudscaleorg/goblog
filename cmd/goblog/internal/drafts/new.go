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
	"path/filepath"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/ldelossa/goblog"
	"gopkg.in/yaml.v3"
)

var newFS = flag.NewFlagSet("new", flag.ExitOnError)

var newFlags = struct {
}{}

func new(ctx context.Context) {
	newFS.Usage = func() {
		fmt.Printf(`
The new subcommand creates a new draft and opens your $EDITOR to it. 

You must save the contents before closing your $EDITOR for GoBlog to correctly safe the draft contents.

On close of the $EDITOR you will choose to either publish or leave the draft for later editing.

This subcommand takes no arguments.

Usage:
	goblog drafts new
`)
	}
	// 0: goblog, 1: posts, 2: edit
	newFS.Parse(os.Args[3:])

	editor := os.Getenv("EDITOR")
	if editor == "" {
		color.Red("Please set your EDITOR environment variable.")
		os.Exit(1)
	}

	scanner := bufio.NewScanner(os.Stdin)
	var draft goblog.Post

	// title prompt
	color.Yellow(`
What's the title of this post?

`)
	fmt.Printf("> ")
	scanner.Scan()
	if err := scanner.Err(); err != nil {
		color.Red("Error: something went wrong inputing your title: %v", err)
		os.Exit(1)
	}
	draft.Title = scanner.Text()

	// summary prompt
	color.Yellow(`
What's the summary of this post?

`)
	fmt.Printf("> ")
	scanner.Scan()
	if err := scanner.Err(); err != nil {
		color.Red("Error: something went wrong inputing your summary: %v", err)
		os.Exit(1)
	}
	draft.Summary = scanner.Text()

	// hero prompt
	color.Yellow(`
If this post will have a hero image you can specify this now. 

Hero images live in the /posts directory so supply a path such as:

"/posts/myposthero.png"

`)
	fmt.Printf("> ")
	scanner.Scan()
	if err := scanner.Err(); err != nil {
		color.Red("Error: something went wrong inputing your summary: %v", err)
		os.Exit(1)
	}
	draft.Hero = scanner.Text()

	if _, err := os.Stat(goblog.Drafts); os.IsNotExist(err) {
		err := os.Mkdir(goblog.Drafts, 0770)
		if err != nil {
			color.Red("Error: failed to create drafts directory: %v", err)
			os.Exit(1)
		}
	}

	formated := strings.ReplaceAll(draft.Title, " ", "_")
	formated = strings.ToLower(formated)
	mdDraft := path.Join(goblog.Drafts, formated+".md")

	// call editor
	var cmd *exec.Cmd
	cmd = exec.Command(editor, mdDraft)
	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin
	err := cmd.Run()
	if err != nil {
		color.Red("Error: failed to start editor: %v", err)
		os.Exit(1)
	}

	// read markdown file into a buffer, close fd and remove
	// the draft.
	f, err := os.OpenFile(mdDraft, os.O_CREATE|os.O_RDWR, 0o0660)
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

	var postPath string

	switch publish {
	case "yes":
		postPath = path.Join(goblog.Posts, formated+".post")
	default:
		postPath = path.Join(goblog.Drafts, formated+".post")
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
Your draft has been written to: %v

It will now be available for usage in subsequent 'drafts' commands.

`, postPath)
	os.Exit(0)
}

// walks the "drafts" directory for drafts posts, sorts them by date, and returns
// a list of them.
func sortedDrafts(ctx context.Context) (goblog.DateSortable, error) {
	var sorted goblog.DateSortable
	err := filepath.Walk(goblog.Drafts, func(path string, info os.FileInfo, err error) error {
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
		post.Path = path
		sorted = append(sorted, post)
		return nil
	})
	if err != nil {
		return sorted, err
	}
	return sorted, nil
}
