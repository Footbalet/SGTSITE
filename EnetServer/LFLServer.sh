#!/bin/sh
printf '\033c\033]0;%s\a' LFLServer
base_path="$(dirname "$(realpath "$0")")"
"$base_path/LFLServer.x86_64" "$@"
