// Package pkgcloud allows you to talk to the packagecloud API.
// See https://packagecloud.io/docs/api
package pkgcloudlib

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"os/user"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/go-errors/errors"

	"github.com/edwarnicke/pkgcloud/pkgcloudlib/upload"
	"github.com/tomnomnom/linkheader"
)

//go:generate bash -c "./gendistros.py supportedDistros | gofmt > distros.go"

// ServiceURL is the URL of packagecloud's API.
const ServiceURL = "https://packagecloud.io/api/v1"

// ServiceBaseURL - base URL for packagecloud
const ServiceBaseURL = "https://packagecloud.io/"

// UserAgent identifier
const UserAgent = "pkgcloud Go client"

// A Client is a packagecloud client.
type Client struct {
	URL   string `json:"url"`
	Token string `json:"token"`
}

// NewClient creates a packagecloud client. API requests are authenticated
// using an API token. If no token is passed, it will be read from the
// PACKAGECLOUD_TOKEN environment variable.
func NewClient(token string) (*Client, error) {
	if token == "" {
		token = os.Getenv("PACKAGECLOUD_TOKEN")
		if token == "" {
			usr, err := user.Current()
			if err != nil {
				return nil, err
			}
			pkfile := filepath.Join(usr.HomeDir, ".packagecloud")
			if _, err := os.Stat(pkfile); err == nil {
				fd, err := os.Open(pkfile)
				if err != nil {
					return nil, err
				}
				client := &Client{}
				err = json.NewDecoder(fd).Decode(client)
				if err != nil {
					return nil, err
				}
				return client, nil
			}
			return nil, errors.New("PACKAGECLOUD_TOKEN unset")
		}
	}
	return &Client{ServiceBaseURL, token}, nil
}

// decodeResponse checks http status code and tries to decode json body
func decodeResponse(resp *http.Response, respJSON interface{}) error {
	switch resp.StatusCode {
	case http.StatusOK, http.StatusCreated:
		return json.NewDecoder(resp.Body).Decode(respJSON)
	case http.StatusUnauthorized, http.StatusNotFound:
		return fmt.Errorf("HTTP status: %s", http.StatusText(resp.StatusCode))
	case 422: // Unprocessable Entity
		var v map[string][]string
		if err := json.NewDecoder(resp.Body).Decode(&v); err != nil {
			return err
		}
		for _, messages := range v {
			for _, msg := range messages {
				// Only return the very first error message
				return errors.New(msg)
			}
			break
		}
		return fmt.Errorf("invalid HTTP body")
	default:
		return fmt.Errorf("unexpected HTTP status: %d", resp.StatusCode)
	}
}

// CreatePackage pushes a new package to packagecloud.
func (c Client) CreatePackage(repo, distro, pkgFile string) error {
	var extraParams map[string]string
	if distro != "" {
		supportedDistros, err := c.SupportedDistros()
		if err != nil {
			return err
		}
		distID, ok := supportedDistros[distro]
		if !ok {
			return fmt.Errorf("invalid distro name: %s", distro)
		}
		extraParams = map[string]string{
			"package[distro_version_id]": strconv.Itoa(distID),
		}
	}

	endpoint := fmt.Sprintf("%s/repos/%s/packages.json", ServiceURL, repo)
	request, err := upload.NewRequest(endpoint, extraParams, "package[package_file]", pkgFile)
	if err != nil {
		return err
	}
	request.SetBasicAuth(c.Token, "")
	request.Header.Add("User-Agent", UserAgent)

	client := &http.Client{}
	resp, err := client.Do(request)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return decodeResponse(resp, &struct{}{})
}

// Package - packagcloud.io Package structure
// See for detailed description of fields: https://packagecloud.io/docs/api#object_PackageFragment
type Package struct {
	Name               string    `json:"name"`
	CreatedAt          time.Time `json:"created_at,string"`
	Epoch              int       `json:"epoch"`
	Scope              string    `json:"scope"`
	Private            bool      `json:"private"`
	UploaderName       string    `json:"uploader_name"`
	Indexed            bool      `json:"indexed"`
	RepositoryHTMLURL  string    `json:"repository_html_url"`
	DownloadDetailsURL string    `json:"downloads_detail_url"`
	DownloadSeriesURL  string    `json:"downloads_series_url"`
	DownloadCountURL   string    `json:"downloads_count_url"`
	PromoteURL         string    `json:"promote_url"`
	DestroyURL         string    `json:"destroy_url"`
	Filename           string    `json:"filename"`
	DistroVersion      string    `json:"distro_version"`
	Version            string    `json:"version"`
	Release            string    `json:"release"`
	Type               string    `json:"type"`
	PackageURL         string    `json:"package_url"`
	PackageHTMLURL     string    `json:"package_html_url"`
}

// Destroy removes package from repository.
//
// repo should be full path to repository
// (e.g. youruser/repository/ubuntu/xenial).
func (c Client) Destroy(repo, packageFilename string) error {
	endpoint := fmt.Sprintf("%s/repos/%s/%s", ServiceURL, repo, packageFilename)

	req, err := http.NewRequest("DELETE", endpoint, nil)
	if err != nil {
		return err
	}

	req.SetBasicAuth(c.Token, "")
	req.Header.Add("User-Agent", UserAgent)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return decodeResponse(resp, &struct{}{})
}

// DestroyFromPackage removes package from repository.
//
// For use with Package struct
func (c Client) DestroyFromPackage(p *Package) error {
	endpoint := fmt.Sprintf("%s/%s", ServiceBaseURL, p.DestroyURL)

	req, err := http.NewRequest("DELETE", endpoint, nil)
	if err != nil {
		return err
	}

	req.SetBasicAuth(c.Token, "")
	req.Header.Add("User-Agent", UserAgent)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return decodeResponse(resp, &struct{}{})
}

// Paginated captures pagination information described at - https://packagecloud.io/docs/api#pagination
type Paginated struct {
	Total      int
	PerPage    int
	MaxPerPage int
}

// PaginatedPackages captures 'Package' and pagination together
// Packages - list of packages returned in this page
// Next - function that can be called to fetch the nexts set of pages
// Paginated - Pagination meta data about this set of packages
type PaginatedPackages struct {
	Packages []*Package
	Next     func() (*PaginatedPackages, error)
	Paginated
}

// GetPaginatedPackages - Gets the first set of PaginatedPackages for endpoint
// Note: Fetching subsequent packages should be done with PackaginesPackages.Next()
func (c *Client) GetPaginatedPackages(endpoint string) (*PaginatedPackages, error) {
	rv := &PaginatedPackages{}
	req, err := http.NewRequest("GET", endpoint, nil)
	if err != nil {
		return nil, err
	}
	req.SetBasicAuth(c.Token, "")
	req.Header.Add("User-Agent", UserAgent)
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	err = decodeResponse(resp, &rv.Packages)
	if err != nil {
		return nil, err
	}
	err = ExtractPaginationHeaders(&resp.Header, &rv.Paginated)
	if err != nil {
		return nil, err
	}
	header := resp.Header.Get("Link")
	links := linkheader.Parse(header)
	for _, link := range links {
		if link.Rel == "next" {
			rv.Next = func() (*PaginatedPackages, error) {
				return c.GetPaginatedPackages(link.URL)
			}
		}
	}
	return rv, nil
}

// ExtractPaginationHeaders - Extract Paginated Object from the http.Headers
func ExtractPaginationHeaders(h *http.Header, p *Paginated) error {
	header := h.Get("Total")
	total, err := strconv.Atoi(header)
	if err != nil {
		return err
	}
	p.Total = total

	header = h.Get("Per-Page")
	perPage, err := strconv.Atoi(header)
	if err != nil {
		return err
	}
	p.PerPage = perPage

	header = h.Get("Max-Per-Page")
	maxPerPage, err := strconv.Atoi(header)
	if err != nil {
		return err
	}
	p.MaxPerPage = maxPerPage
	return nil
}

// PaginatedAll - Get the list of all Packages from a repo using PaginatedPackages
// The first PaginatedPackages object is the first page of responses.
// To get subsequent pages, call PaginatedPackages.Next() if it is non-nil
func (c *Client) PaginatedAll(repo string) (*PaginatedPackages, error) {
	endpoint := fmt.Sprintf("%s/repos/%s/packages.json", ServiceURL, repo)
	return c.GetPaginatedPackages(endpoint)
}

// Promote - Promote Package to repo
func (c *Client) Promote(p *Package, repo string) error {
	endpoint := fmt.Sprintf("%s/%s", ServiceBaseURL, p.PromoteURL)
	form := url.Values{}
	form.Add("destination", repo)
	req, err := http.NewRequest("POST", endpoint, strings.NewReader(form.Encode()))
	if err != nil {
		return err
	}
	req.SetBasicAuth(c.Token, "")
	req.Header.Add("User-Agent", UserAgent)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return decodeResponse(resp, &struct{}{})
}

// Distributions - struct to represent how packagecloud.io handles distributions
// https://packagecloud.io/docs/api#resource_distributions
type Distributions struct {
	Deb []Distribution `json:"deb"`
	Dsc []Distribution `json:"dsc"`
	Rpm []Distribution `json:"rpm"`
}

// Distribution - struct to represent how packagecloud.io handles distributions
// https://packagecloud.io/docs/api#resource_distributions
type Distribution struct {
	DisplayName string `json:"display_name"`
	IndexName   string `json:"index_name"`
	Versions    []DistributionVersion
}

// DistributionVersion - struct to represent how packagecloud.io handles distributions
// https://packagecloud.io/docs/api#resource_distributions
type DistributionVersion struct {
	ID          int    `json:"id"`
	DisplayName string `json:"display_name"`
	IndexName   string `json:"index_name"`
}

// Distributions - retrieve all distribution descriptions
func (c *Client) Distributions() (*Distributions, error) {
	endpoint := fmt.Sprintf("%s/%s", ServiceURL, "distributions.json")
	req, err := http.NewRequest("GET", endpoint, nil)
	if err != nil {
		return nil, err
	}
	req.SetBasicAuth(c.Token, "")
	req.Header.Add("User-Agent", UserAgent)
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	distributions := &Distributions{}
	err = decodeResponse(resp, distributions)
	if err != nil {
		return nil, err
	}
	return distributions, nil
}

// SupportedDistros - return a map of distro strings like "ubuntu/xenial" to distro ids.
func (c *Client) SupportedDistros() (map[string]int, error) {
	rv := make(map[string]int, 256)
	d, err := c.Distributions()
	if err != nil {
		return nil, err
	}
	for _, dist := range d.Deb {
		for _, v := range dist.Versions {
			rv[fmt.Sprintf("%s/%s", dist.IndexName, v.IndexName)] = v.ID
		}
	}
	return rv, nil
}

// Exists - Check to see if <repo>/<distro>/packageFilename exists in packagecloud.io
func (c *Client) Exists(repo, distro, packageFilename string) (bool, error) {
	endpoint := fmt.Sprintf("%s/%s/packages/%s/%s", ServiceBaseURL, repo, distro, packageFilename)

	req, err := http.NewRequest("HEAD", endpoint, nil)
	if err != nil {
		return false, err
	}

	req.SetBasicAuth(c.Token, "")
	req.Header.Add("User-Agent", UserAgent)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		if err.Error() == "HTTP status: Not Found" {
			return false, nil
		}
		return false, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return false, nil
	}
	return true, nil
}
