package config

import (
	"context"
	"flag"
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/fatih/color"
	"github.com/ldelossa/goblog"
	"gopkg.in/yaml.v2"
)

var appPathsFS = flag.NewFlagSet("app-paths", flag.ExitOnError)

var appPathsFlags = struct{}{}

func appPaths(ctx context.Context) {
	color.Blue(`
Provide a comma separated list of paths your front-end web application will handle.

Each path should have a leading forward slash.

Example: /post,/archive,/settings

`)

	var list string
	_, err := fmt.Scanln(&list)
	if err != nil {
		color.Red("failed to scan input: %v", err)
		os.Exit(1)
	}

	paths := strings.Split(list, ",")
	color.Blue("Adding the following paths: %v\n", paths)

	goblog.Conf.AppPaths = paths
	dest := path.Join(goblog.Home, "src/config/config.yaml")
	if _, err := os.Stat(dest); err != nil {
		color.Red("Could not stat config: %v", err)
		os.Exit(1)
	}
	f, err := os.OpenFile(dest, os.O_RDWR|os.O_TRUNC, 0)
	if err != nil {
		color.Red("Failed to open config.yaml: %v", err)
		os.Exit(1)
	}
	defer f.Close()
	if err = yaml.NewEncoder(f).Encode(&goblog.Conf); err != nil {
		color.Red("Failed to write config: %v", err)
		os.Exit(1)
	}
	color.Blue("Wrote new config to %v\n", dest)
}
