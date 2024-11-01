package main

import "time"

func processMarkdownFile(
	file MarkdownFile,
	jekyllDir *string,
	zolaDir *string,
	tz *time.Location,
) error {
	if err := file.Load(); err != nil {
		return err
	}
	if err := file.ProcessFrontMatter(); err != nil {
		return err
	}
	if err := file.ConvertToTOML(tz); err != nil {
		return err
	}
	return file.Save(jekyllDir, zolaDir)
}
