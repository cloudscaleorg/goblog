package goblog

import (
	"embed"

	"gopkg.in/yaml.v3"
)

var Conf Config

func init() {
	var conf Config
	f, err := ConfigFS.Open("config/config.yaml")
	if err != nil {
		panic("could not open config: " + err.Error())
	}

	err = yaml.NewDecoder(f).Decode(&conf)
	if err != nil {
		panic("could not json decode config: " + err.Error())
	}
	Conf = conf
}

//go:embed config
var ConfigFS embed.FS
