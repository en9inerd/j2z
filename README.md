# j2z

Current status: **Early development**

## Expected usage and functionality

The goal of this project is to provide a simple tool to convert a Jekyll markdown posts to Zola markdown posts.

Usage:

```sh
j2z --jekyllDir <path> --zolaDir <path> [--timezone <timezone>] [--taxomonies <taxomonies>]
```

Where:
- `--jekyllDir` is the path to the Jekyll directory containing the `_config.yml` file.
- `--zolaDir` is the path to the Zola directory containing the `content` directory.
- `--timezone` (optional) is the timezone to use for the conversion. If not provided, the timezone will be taken from local machine. Example: `America/New_York`.
- `--taxomonies` (optional) is the taxomonies to use for the conversion. Example: `tags,categories`.
