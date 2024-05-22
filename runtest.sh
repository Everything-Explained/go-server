#!/bin/bash

# Check if gotestsum is available
if ! command -v gotestsum >/dev/null 2>&1; then
    gotestsum -f testdox --watch ./...
else
    echo -e "\n\033[1m\033[31mMissing Go Module: \033[0mhttps://github.com/gotestyourself/gotestsum"
fi
