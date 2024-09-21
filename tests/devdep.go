//go:build gomoddep

package main

import (
	_ "github.com/thehowl/tomino/cmd/tomgen"
	_ "github.com/thehowl/tomino/generator"
	_ "mvdan.cc/gofumpt"
)
