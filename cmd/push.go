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
	"log"
	"os"
	"path/filepath"
	"strings"

	pkgcloud "github.com/edwarnicke/pkgcloud/pkgcloudlib"
	"github.com/spf13/cobra"
)

var pushCmd = &cobra.Command{
	Use:   "push user/repo/distro/version/ filename",
	Short: "push package to repo",
	Long:  `push package to repo`,
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
			if _, err := os.Stat(args[i]); os.IsNotExist(err) {
				log.Fatalf("%s does not exist", args[i])
			}
		}
		for i := 1; i < len(args); i++ {
			repodistro := fmt.Sprintf("%s/%s", repo, distro)
			path := args[i]
			filename := filepath.Base(path)
			if !DryRun {
				exists, err := client.Exists(repo, distro, filename)
				if err != nil {
					log.Fatalf("error: %s\n", err)
				}
				if exists {
					if !force {
						log.Fatalf("package %s already exists in repo %s/%s, use -f to force overwrite", filename, repo, distro)
					}
					log.Printf("package %s already exists in repo %s/%s. -f provided.  Deleting in preparation to push new version", filename, repo, distro)
					err = client.Destroy(repodistro, filename)
					if err != nil {
						log.Fatalf("error deleting %s from %s in preparation for overwrite: %s\n", filename, repodistro, err)
					}
				}
				err = client.CreatePackage(repo, distro, path)
				if err != nil {
					log.Fatalf("error: %s\n", err)
				}
				log.Printf("Pushed %s to %s", path, repodistro)
			} else {
				exists, err := client.Exists(repo, distro, filename)
				if err != nil {
					log.Fatalf("error: %s\n", err)
				}
				if exists {
					if !force {
						log.Fatalf("Dry Run package %s already exists in repo %s/%s, use -f to force overwrite", filename, repo, distro)
					}
					log.Printf("Dry Run package %s already exists in repo %s/%s. -f provided.  Deleting in preparation to push new version", filename, repo, distro)
				}
				log.Printf("Dry Run for pushing %s to %s", args[i], repodistro)
			}
		}
	},
	Args:             cobra.MinimumNArgs(2),
	TraverseChildren: true,
}

var force bool

func init() {
	pushCmd.Flags().BoolVarP(&force, "force", "f", false, "Force overwrite of package if it already exists")
}
