// Copyright (c) 2018 Joseph Benden <joe@benden.us>
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
	"log"
	"strings"

	pkgcloud "github.com/jbenden/pkgcloud/pkgcloudlib"
	"github.com/spf13/cobra"
)

var destroyCmd = &cobra.Command{
	Use:   "destroy user/repo/distro/version/ filename",
	Short: "destroy a package",
	Long:  `destroy/remove a package from the repo`,
	Run: func(cmd *cobra.Command, args []string) {
		parts := strings.Split(args[0], "/")
		if len(parts) != 4 {
			log.Fatalf("%s is not of form user/repo/distro/version/", args[0])
		}
		repo := parts[0] + "/" + parts[1]
		distro := parts[2] + "/" + parts[3]
		client, err := pkgcloud.NewClient("")
		if err != nil {
			log.Fatalf("error: %s\n", err)
		}
		for i := 1; i < len(args); i++ {
			repodistro := fmt.Sprintf("%s/%s", repo, distro)
			filename := args[i]

			exists, err := client.Exists(repo, distro, filename)
			if err != nil {
				log.Fatalf("error: %s\n", err)
			}

            if exists {
                if !DryRun {
					err = client.Destroy(repodistro, filename)
					if err != nil {
						log.Fatalf("error deleting %s from %s in preparation for overwrite: %s\n", filename, repodistro, err)
					}
					log.Printf("Destroyed %s on %s", filename, repodistro)
                } else {
                    log.Printf("Dry Run for destroying %s from %s", args[i], repodistro)
				}
			}
		}
	},
	Args:             cobra.MinimumNArgs(2),
	TraverseChildren: true,
}
