package args

import "time"

type Args struct {
	JekyllDir     string
	ZolaDir       string
	Taxonomies    []string
	ExtraRootKeys []string
	Tz            *time.Location
	Aliases       bool
	DryRun        bool
}
