package initialize

import (
	"bufio"
	"context"
	"errors"
	"flag"
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"path"

	"github.com/fatih/color"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/ldelossa/goblog"
	"github.com/ldelossa/goblog/pkg/dtree"
	"gopkg.in/yaml.v3"
)

var initFS = flag.NewFlagSet("init", flag.ExitOnError)

var initFlags = struct{}{}

func Initialize(ctx context.Context) {
	root := buildDTree()
	err := root.Execute(ctx)
	if err != nil {
		color.Red("Error: %v\n", err)
		os.Exit(1)
	}
}

// Initialization is driven off whether a goblog home directory
// can be found.
//
// If the home directory cannot be found the user will be asked to
// define one, provide a remote and branch to clone the goblog source
// into this home directory, and build a new goblog binary.
//
// If a home directory is found we check to see if a git repo
// exists in it.
//
//				  Decision Tree for Initialization
//						 [HomeExistsDecision]
//						 no/            \yes
//	      [GitCloneDecision]            [CheckGitRepoDecision]
//	      /         \yes                         no/ \yes
//      [nil]   [BuildDecision]   [GitCloneDecision][BuildNumDecision]
//       /\            /\                 no/ \yes               no/ \yes
//  [nil]  [nil]  [nil] [nil]           [nil] [BuildDecision] [nil]  [nil]
func buildDTree() *dtree.Decision {
	// representation of binary tree
	// as array.
	nodes := [...]*dtree.Decision{
		NewHomeExistsDecision(),
		NewGitCloneDecision(),
		NewCheckGitRepoDecision(),
		nil,
		NewBuildDecision(),
		NewGitCloneDecision(),
		NewBuildNumDecision(),
		nil,
		nil,
		nil,
		nil,
		nil,
		NewBuildDecision(),
		nil,
		nil,
	}

	for i, node := range nodes {
		if (2*i)+1 > len(nodes)-1 {
			break
		}
		if node == nil {
			continue
		}
		node.AddNo(
			nodes[(2*i)+1],
		)
		node.AddYes(
			nodes[(2*i)+2],
		)
	}
	return nodes[0]
}

// NewHomeExistsDecision determines if GoBlog's home directory
// exists. This home directory is resolved via the user.Current() function
// in the os/user package.
//
// If its home does exist the decision calls its Yes branch and logs its location.
//
// If it does not exist an attempt to create the dir is made.
// On success its No branch will be called indicating a home directory
// did not exist before this Decision.
//
// If an error occurs creating the directory an error value is returned.
func NewHomeExistsDecision() *dtree.Decision {
	return &dtree.Decision{
		Exec: func(ctx context.Context) (bool, error) {
			fi, err := os.Stat(goblog.Home)
			pathErr := new(os.PathError)
			switch {
			case err == nil:
				if !fi.IsDir() {
					return false, fmt.Errorf("Looks like you have a regular file named goblog in your home dir: %v. You'll need to remove this before GoBlog can continue.", goblog.Home)
				}
				color.Blue(`
GoBlog found its home @ %s

`, goblog.Home)
				return true, nil
			case errors.As(err, &pathErr):
			default:
				return false, err
			}

			color.Yellow(`
GoBlog could not find its home directory. 

Don't worry we will create it for you, but first lets explain how GoBlog works.

GoBlog utilizes Go's embed fs features to statically embed and serve all your blog contents.

In order to do this GoBlog must keep its own source code aronnd and rebuild itself when you add or remove contents.

Obviously you'd like your blog content in a repo of your own, not in GoBlog's upstream, so you should now make a fork of https://github.com/ldelossa/goblog.git if you haven't yet.

When adding or removing contents these changes will be pushed to your fork, you will soon be asked to provide the http remote of your fork and a branch GoBlog will push changes too. 

`)
			color.Blue("Making GoBlog home directory %v\n", goblog.Home)
			err = os.Mkdir(goblog.Home, 0o750)
			if err != nil {
				return false, fmt.Errorf("Error creating GoBlog home: %w", err)
			}
			return false, nil
		},
	}
}

// GetGitCloneDecision will attempt to clone GoBlog from the provided
// git remote and branch.
//
// This Decision always calls its Yes branch.
//
// If an error is encountered during or up to the clone an error is returned.
func NewGitCloneDecision() *dtree.Decision {
	return &dtree.Decision{
		Exec: func(ctx context.Context) (bool, error) {
			remote, branch := goblog.Conf.Remote, goblog.Conf.Branch
			_, err := url.Parse(remote)
			if err != nil {
				return false, fmt.Errorf("Could not parse remote URL. GoBlog only allows cloning from public http repositories currently: %w", err)
			}
			for remote == "" || branch == "" {

				color.Yellow(`
GoBlog needs a fork to push your blog content to. 

If you haven't yet please fork https://github.com/ldelossa/goblog.git and provide your fork's http remote.

`)
				fmt.Printf("> ")
				_, err := fmt.Scanln(&remote)
				if err != nil {
					return false, err
				}
				color.Yellow(`
Now, what branch do you want to checkout? 

When adding blog with GoBlog you'll push these changes to this branch. (default: master)

`)
				fmt.Printf("> ")
				scnr := bufio.NewScanner(os.Stdin)
				scnr.Scan()
				if scnr.Err() != nil {
					color.Red("Error: failed to scan branch: %v", scnr.Err())
					os.Exit(1)
				}
				if scnr.Text() == "" {
					branch = "master"
				} else {
					branch = scnr.Text()
				}
			}

			goblog.Conf.Remote, goblog.Conf.Branch = remote, branch

			src := path.Join(goblog.Home, "src")
			color.Blue("Cloning %v and checking out branch %v into %v\n\n", remote, branch, src)

			// clone user's goblog fork
			repo, err := git.PlainCloneContext(ctx, src, false, &git.CloneOptions{
				URL:           "https://github.com/ldelossa/goblog.git",
				ReferenceName: "master",
			})
			if err != nil {
				return false, fmt.Errorf("Failed to clone remote %v and checkout %v: %w", remote, branch, err)
			}

			err = repo.Fetch(&git.FetchOptions{
				Tags:       git.AllTags,
				RemoteName: "origin",
			})

			repo.CreateRemote(&config.RemoteConfig{
				Name: "fork",
				URLs: []string{
					remote,
				},
			})

			return true, nil
		},
	}
}

// NewCheckGitRepoDecision returns a Decision which
// determines if GoBlog's source code exist in GoBlog's
// home directory.
//
// If it does it calls its Yes branch, if not it
// calls its No branch.
func NewCheckGitRepoDecision() *dtree.Decision {
	return &dtree.Decision{
		Exec: func(ctx context.Context) (bool, error) {
			src := path.Join(goblog.Home, "src")
			_, err := git.PlainOpen(src)
			switch {
			case err == git.ErrRepositoryNotExists:
				color.Blue("Can't seem to find the GoBlog source code in your home.\n\n")
				return false, nil
			case err != nil:
				return false, fmt.Errorf("Failed to check if GoBlog source code is available: %w", err)
			}
			return true, nil
		},
	}
}

// NewBuildDecision returns a Decision which builds
// a new goblog binary.
//
// This decision always calls its Yes branch or errors.
func NewBuildDecision() *dtree.Decision {
	return &dtree.Decision{
		Exec: func(ctx context.Context) (bool, error) {
			// bump the build version
			goblog.Conf.BuildNum++

			// write out the new config
			dest := path.Join(goblog.Home, "src/config/config.yaml")
			if _, err := os.Stat(dest); err != nil {
				return false, fmt.Errorf("Could not stat config: %v", err)
			}
			f, err := os.OpenFile(dest, os.O_RDWR|os.O_TRUNC, 0)
			if err != nil {
				return false, fmt.Errorf("Failed to open config.yaml: %v", err)
			}
			defer f.Close()
			if err = yaml.NewEncoder(f).Encode(&goblog.Conf); err != nil {
				return false, fmt.Errorf("Failed to write config: %v", err)
			}
			color.Blue(`
Wrote new config to %v

`, dest)

			// build new binary
			buildDest := path.Join(goblog.Home, "bin")
			_, err = os.Stat(buildDest)
			switch {
			case os.IsNotExist(err):
				if err := os.Mkdir(buildDest, 0o750); err != nil {
					return false, fmt.Errorf("Failed creating bin directory: %w", err)
				}
			case err != nil:
				return false, fmt.Errorf("Failed to stat build directory: %v", err)
			default:
			}

			goPath, err := exec.LookPath("go")
			if err != nil {
				return false, fmt.Errorf("could not find go command in path: %w", err)
			}
			color.Blue(`
Building your new GoBlog binary @ %v

`, buildDest)
			goBuild := exec.Cmd{
				Path:   goPath,
				Args:   []string{"go", "build", "-o", "../bin/goblog", "./cmd/goblog"},
				Dir:    path.Join(goblog.Home, "src"),
				Stdin:  os.Stdin,
				Stdout: os.Stdout,
				Stderr: os.Stderr,
			}
			err = goBuild.Run()
			if err != nil {
				return false, fmt.Errorf("Failed to build GoBlog: %v", err)
			}
			color.Blue(`
A new GoBlog binary has been successfuly built @ %s/%s

You'll want to give this binary a test drive with 'goblog serve' and make sure your posts look good.

If all looks well you can push your changes to your fork and discard the current binary.

If things look off, git reset your GoBlog's src directory and your current binary will begin to work once again.

`, buildDest, "goblog")
			return true, nil
		},
	}
}

// NewBuildNumDecision returns a Decision which
// returns an error if the embedded config build number
// does not match the build number in the goblog's home.
func NewBuildNumDecision() *dtree.Decision {
	return &dtree.Decision{
		Exec: func(ctx context.Context) (bool, error) {
			// allow a trap door here for testing.
			if os.Getenv("GOBLOG_BUILDCHECK") != "" {
				return true, nil
			}

			var srcConf goblog.Config
			sourceConfPath := path.Join(goblog.Home, "src/config/config.yaml")
			f, err := os.Open(sourceConfPath)
			if err != nil {
				return false, err
			}
			err = yaml.NewDecoder(f).Decode(&srcConf)
			if err != nil {
				return false, err
			}
			if srcConf.BuildNum != goblog.Conf.BuildNum {
				return false, fmt.Errorf(`
The GoBlog binary you are using does not match the config in your src directory (%s).

Either reset your src directory or use the latest GoBlog binary in your GoBlog home's bin directory. 

If you've misplaced your latest GoBlog binary you can use the 'publish' subcommand to build a new one.

If you know what you're doing you can run the command with the env variable "GOBLOG_BUILDCHECK=false" to bypass this check.
`, sourceConfPath)
			}
			return true, nil
		},
	}
}
