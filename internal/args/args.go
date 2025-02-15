package args

import "time"

// struct to hold all arguments
type Args struct {
	JekyllDir  string
	ZolaDir    string
	Taxonomies []string
	Tz         *time.Location
	Aliases    bool
}
