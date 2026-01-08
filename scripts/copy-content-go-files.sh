#!/bin/bash

find . -name "*.go" -type f -exec sh -c 'printf "# File: %s\n\n" "$1"; cat "$1"; printf "\n#############################\n"' _ {} \; | xclip -selection clipboard
