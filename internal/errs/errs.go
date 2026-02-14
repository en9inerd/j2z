package errs

import "fmt"

// FrontMatterError represents an error parsing or processing front matter.
type FrontMatterError struct {
	File string
	Msg  string
	Err  error
}

func (e *FrontMatterError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("front matter error in %s: %s: %v", e.File, e.Msg, e.Err)
	}
	return fmt.Sprintf("front matter error in %s: %s", e.File, e.Msg)
}

func (e *FrontMatterError) Unwrap() error { return e.Err }

// FilenameError represents an error parsing a Jekyll filename.
type FilenameError struct {
	Name string
	Msg  string
}

func (e *FilenameError) Error() string {
	return fmt.Sprintf("filename error for %q: %s", e.Name, e.Msg)
}

// DateError represents an error parsing a date in front matter.
type DateError struct {
	File   string
	Value  string
	Reason string
}

func (e *DateError) Error() string {
	return fmt.Sprintf("date error in %s: could not parse %q: %s", e.File, e.Value, e.Reason)
}
