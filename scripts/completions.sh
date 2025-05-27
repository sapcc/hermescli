#!/bin/sh

# SPDX-FileCopyrightText: 2025 SAP SE or an SAP affiliate company
#
# SPDX-License-Identifier: Apache-2.0

set -e
rm -rf completions
mkdir completions
for sh in bash zsh fish; do
	go run cmd/main.go completion "$sh" >"completions/hermescli.$sh"
done
