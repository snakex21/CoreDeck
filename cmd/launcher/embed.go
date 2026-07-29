///usr/bin/env go run "$0" "$@"; exit
package main

import "embed"

//go:embed all:web
var staticFiles embed.FS
