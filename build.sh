#!/bin/bash

platforms=("linux/amd64" "linux/arm64" "darwin/amd64" "darwin/arm64" "windows/amd64" "windows/arm64")

for platform in "${platforms[@]}"
do
  GOOS="${platform%/*}"
  GOARCH="${platform#*/}"
  output="dist/$(basename $PWD)_${GOOS}_${GOARCH}"

  if [ "$GOOS" == "windows" ]; then
    output="$output.exe"
  fi

  echo "Building $output"
  GOOS=$GOOS GOARCH=$GOARCH go build -gcflags="all=-l -B" -trimpath -ldflags="-s -w" -o $output ./src
done
