# j2z

The goal of this project is to provide a simple tool to convert a Jekyll markdown posts to Zola markdown posts.

## Usage:

```sh
j2z --jekyllDir <path> --zolaDir <path> [--tz <timezone>] [--taxonomies <taxonomies>] [--aliases <true|false>]
```

## Flags:
- `--jekyllDir` (required): Specifies the path to the Jekyll directory containing the` _config.yml` file.
- `--zolaDir` (required): Specifies the path to the Zola directory containing the `content` directory.
- `--tz` (optional): Sets the timezone for the conversion. If not provided, the timezone will default to the local machine's timezone. Example: `America/New_York`.
- `--taxonomies` (optional): A comma-separated list of taxonomies to include in the conversion. Default is `tags,categories`.
- `--aliases` (optional): Enables aliases in the front matter if set to `true`. Default is `true`.
