package args

import (
	"os"
	"time"
)

type Args struct {
	JekyllDir     string
	ZolaDir       string
	Taxonomies    []string
	ExtraRootKeys []string
	Tz            *time.Location
	Aliases       bool
	DryRun        bool
	OutputRoot    *os.Root
}
