package processor

import (
	"github.com/en9inerd/j2z/internal/args"
	"github.com/en9inerd/j2z/internal/file"
)

func ProcessMarkdownFile(
	file file.MarkdownFile,
	args *args.Args,
) error {
	if err := file.Load(); err != nil {
		return err
	}
	if err := file.ProcessFrontMatter(); err != nil {
		return err
	}
	if err := file.ConvertToTOML(args); err != nil {
		return err
	}
	return file.Save(args)
}
