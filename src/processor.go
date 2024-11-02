package main

func processMarkdownFile(
	file MarkdownFile,
	args *Args,
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
