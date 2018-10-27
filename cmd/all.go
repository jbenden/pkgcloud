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
	"text/template"
	"log"
	"os"
	"time"

	pkgcloud "github.com/jbenden/pkgcloud/pkgcloudlib"
	"github.com/spf13/cobra"
)

var allCmd = &cobra.Command{
	Use:   "all <user/repo>",
	Short: "List all the packages in a repo",
	Long:  `List all the packages in a repo`,
	Run: func(cmd *cobra.Command, args []string) {
		repo := args[0]
		client, err := pkgcloud.NewClient("")
		if err != nil {
			log.Fatalf("error: %s\n", err)
		}
		t := template.Must(template.New("package-tmpl").Parse(allTemplateString))

		next := func() (*pkgcloud.PaginatedPackages, error) {
			return client.PaginatedAll(repo)
		}
		var packages []*pkgcloud.Package
		for next != nil {
			paginatedPackages, err := next()
			if err != nil {
				log.Fatalf("pagination error: %s\n", err)
			}
			packages = append(packages, paginatedPackages.Packages...)
			for _, p := range paginatedPackages.Packages {
				pack := &Package{Package: p}
				t.Execute(os.Stdout, pack)
			}
			next = paginatedPackages.Next
		}
		if !DryRun {
			for _, p := range packagesToDestroy {
				err = client.DestroyFromPackage(p.Package)
				if err != nil {
					log.Fatalf("Error when trying to Destroy %s : %s", p.PackageHTMLURL, err)
				}
				log.Printf("Destroying %s\n", p.PackageHTMLURL)
			}
			for p, r := range packagesToPromote {
				err = client.Promote(p.Package, r)
				if err != nil {
					log.Fatalf("Error Promoting to %s : %s : %s", r, p.PromoteURL, err)
				}
				log.Printf("Promoted to %s : %s\n", r, p.PromoteURL)
			}
		}
	},
	Args:             cobra.ExactArgs(1),
	TraverseChildren: true,
}

var allTemplateString string

func init() {
	allCmd.Flags().StringVarP(&allTemplateString, "template", "t", "{{.PackageHTMLURL}}\n", "Golang text template for output")
}

// Package - wraps pkgcloud.Package in order to allow adding 'convenience' method
type Package struct {
	*pkgcloud.Package
}

var packagesToPromote = make(map[*Package]string)

// Promote - promote Package to repo
func (p *Package) Promote(repo string) string {
	packagesToPromote[p] = repo
	return fmt.Sprintf("Marked for Promotion to %s : %s", repo, p.PromoteURL)
}

// DaysOld - Number of days old the Package is
func (p *Package) DaysOld() int {
	return int(time.Since(p.CreatedAt).Hours() / 24)
}

var packagesToDestroy []*Package

// Destroy - destroy the package referenced by pkgcloud.Package
func (p *Package) Destroy() string {
	packagesToDestroy = append(packagesToDestroy, p)
	return fmt.Sprintf("Marked for Destruction %s", p.PackageHTMLURL)
}
