// Copyright (c) 2018 Cisco and/or its affiliates.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at:
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// DryRun - if true, don't do anything that would change packagecloud.io state
var DryRun bool

var rootCmd = &cobra.Command{
	Use:   "pkgcloud",
	Short: "pkgcloud is a command-line for packagecloud.io",
	Long:  `pkgcloud is a command-line for packagecloud.io`,
	Run: func(cmd *cobra.Command, args []string) {
		// Do Stuff Here
	},
	TraverseChildren: true,
}

// Execute the pkgcloud command
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().BoolVarP(&DryRun, "dry-run", "d", false, "Do not take actions that change the state of packagecloud.io")
	rootCmd.AddCommand(allCmd)
	rootCmd.AddCommand(distributionsCmd)
	rootCmd.AddCommand(pushCmd)
}
