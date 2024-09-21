#!/bin/sh
if [ "$1" = "fix" ]; then
    mv result.go.1 result.go
fi

go run github.com/thehowl/tomino/cmd/tomgen \
    net/url.URL > result.go.1 || exit 1

diff --color -su result.go.1 result.go
sc="$?"
if [ "$sc" != "0" ]; then
    exit $sc
else
    rm result.go.1
fi
