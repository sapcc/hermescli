#!/bin/sh
set -e
rm -rf completions
mkdir completions
for sh in bash zsh fish; do
	go run client/main.go completion "$sh" >"completions/hermescli.$sh"
done
