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
	"html/template"
	"log"
	"os"

	pkgcloud "github.com/edwarnicke/pkgcloud/pkgcloudlib"
	"github.com/spf13/cobra"
)

var distributionsCmd = &cobra.Command{
	Use:   "distributions",
	Short: "List all distributions",
	Long:  `List all distributions`,
	Run: func(cmd *cobra.Command, args []string) {
		client, err := pkgcloud.NewClient("")
		if err != nil {
			log.Fatalf("error: %s\n", err)
		}
		distributions, err := client.Distributions()
		if err != nil {
			log.Fatalf("error: %s\n", err)
		}
		t := template.Must(template.New("package-tmpl").Parse(distributionTemplateString))
		dist := &Distributions{Distributions: distributions}
		t.Execute(os.Stdout, dist)
	},
	Args:             cobra.ExactArgs(0),
	TraverseChildren: true,
}

var distributionTemplateString string

func init() {
	distributionsCmd.Flags().StringVarP(&distributionTemplateString, "template", "t", "{{range .Linearize}}{{.DistributionIndex}}/{{.VersionIndex}}: {{.ID}}\n{{end}}", "Golang text template for output")
}

// Distributions - wraps pkgcloud.Distributions to allow adding convenience methods
type Distributions struct {
	*pkgcloud.Distributions
}

// LinearizesDistribution - pkgcloud.Distributions is really not what you are usually looking for
// LinearizesDistribution linearizes it to a more usefule form: one entry per distribution
type LinearizesDistribution struct {
	ID                int
	Type              string
	DistributionName  string
	DistributionIndex string
	VersionName       string
	VersionIndex      string
}

// Linearize - translate a Distributions into a list of LinearizedDistributions
func (d *Distributions) Linearize() []*LinearizesDistribution {
	var rv []*LinearizesDistribution
	for _, dist := range d.Deb {
		for _, v := range dist.Versions {
			rv = append(rv, &LinearizesDistribution{
				ID:                v.ID,
				Type:              "deb",
				DistributionName:  dist.DisplayName,
				DistributionIndex: dist.IndexName,
				VersionName:       v.DisplayName,
				VersionIndex:      v.IndexName,
			})
		}
	}
	for _, dist := range d.Dsc {
		for _, v := range dist.Versions {
			rv = append(rv, &LinearizesDistribution{
				ID:                v.ID,
				Type:              "dsc",
				DistributionName:  dist.DisplayName,
				DistributionIndex: dist.IndexName,
				VersionName:       v.DisplayName,
				VersionIndex:      v.IndexName,
			})
		}
	}
	for _, dist := range d.Rpm {
		for _, v := range dist.Versions {
			rv = append(rv, &LinearizesDistribution{
				ID:                v.ID,
				Type:              "rpm",
				DistributionName:  dist.DisplayName,
				DistributionIndex: dist.IndexName,
				VersionName:       v.DisplayName,
				VersionIndex:      v.IndexName,
			})
		}
	}
	return rv
}
