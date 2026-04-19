# j2z

Simple CLI tool to convert Jekyll markdown posts to Zola markdown posts.

## Installation

### Homebrew (macOS / Linux)

```
brew install en9inerd/tap/j2z
```

### Pre-built binaries

Download from the [Releases](https://github.com/en9inerd/j2z/releases) page.

### From source

```
go install github.com/en9inerd/j2z/cmd/j2z@latest
```

## Usage:

```sh
j2z --jekyll-dir <path> --zola-dir <path> [flags]
```

## Flags:
- `-j, --jekyll-dir` (required): Path to the Jekyll directory containing `_posts` and other underscore-prefixed directories.
- `-z, --zola-dir` (required): Path to the Zola directory where converted files will be written under `content/`.
- `--tz` (optional): Timezone name for date parsing. Defaults to the local machine's timezone. Example: `America/New_York`.
- `--taxonomies` (optional): Comma-separated list of taxonomies to include. Default: `tags,categories`.
- `--extra-root-keys` (optional): Comma-separated list of additional front matter keys to keep at root level (instead of moving to `[extra]`).
- `--aliases` (optional): Enable aliases in the front matter derived from Jekyll filenames.
- `--dry-run` (optional): Preview conversion without writing any files.
- `-v, --verbose` (optional): Enable verbose (debug-level) logging.
- `-q, --quiet` (optional): Suppress all output except errors.
- `--version`: Print version, commit hash, and build time.

## Features:
- Converts YAML front matter to TOML
- Maps Jekyll `last_modified_at` to Zola `updated` field
- Converts `{% highlight lang %}` Liquid tags to fenced code blocks
- Warns on unsupported `{% include %}` Liquid tags
- Normalizes `<!--more-->` summary break tags
- Concurrent file processing with bounded parallelism
- Structured error reporting with typed errors

## Requirements:
- Go 1.26+

## License

MIT
