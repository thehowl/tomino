module hello

go 1.22.2

require github.com/thehowl/tomino v0.0.0-latest

require (
	golang.org/x/mod v0.21.0 // indirect
	golang.org/x/sync v0.8.0 // indirect
	golang.org/x/tools v0.25.0 // indirect
)

replace github.com/thehowl/tomino => ../..
