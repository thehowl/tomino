//go:build gomoddep

package tests

import (
	_ "github.com/thehowl/tomino/cmd/tomgen"
	_ "github.com/thehowl/tomino/generator"
	_ "mvdan.cc/gofumpt"
)
