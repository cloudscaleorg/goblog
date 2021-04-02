package goblog

// Version is the goblog binary version.
//
// This will increment when a new GoBlog is released.
// This supports identifying upgrades.
type Version string

type Config struct {
	// BuildNum allows goblog to determine if its running against
	// a newer home directory. A user should not edit this value
	// manually.
	BuildNum int64
	// Remote tells goblog where its initial repository
	// will be cloned from and where it will optionally
	// push embedfs changes to.
	Remote string `json:"remote" yaml:"remote"`
	// The branch which will be checked out after cloning
	// the Remote.
	//
	// If empty "master" will be assumed.
	Branch string `json:"branch" yaml:"branch"`
	// The paths your front-end web applications serves.
	// When GoBlog encounters these paths it will serve
	// your web application's index.html
	//
	// This is how deep linking is supported.
	AppPaths []string
}
