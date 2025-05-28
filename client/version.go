// SPDX-FileCopyrightText: 2020 SAP SE or an SAP affiliate company
// SPDX-License-Identifier: Apache-2.0

package client

import (
	"fmt"
	"runtime"

	"github.com/spf13/cobra"
)

var Version = "dev"

var VersionCmd = &cobra.Command{
	Use:               "version",
	Short:             "Print version information",
	DisableAutoGenTag: true,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("hermescli %s compiled with %v on %v/%v\n",
			Version, runtime.Version(), runtime.GOOS, runtime.GOARCH)
	},
}

func init() {
	RootCmd.AddCommand(VersionCmd)
}
