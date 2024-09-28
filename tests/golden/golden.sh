#!/bin/sh
# cd into script directory
cd "$(dirname "$0")"

go run github.com/thehowl/tomino/cmd/tomgen \
    net/url.URL \
    github.com/thehowl/tomino/tests/golden.TestType > result.go.1 || exit 1

diff --color -bsu result.go result.go.1
sc="$?"
if [ "$sc" != "0" ]; then
    if [ "$1" = "fix" ]; then
        mv result.go.1 result.go
    fi
    exit $sc
else
    rm result.go.1
fi
