## Installation

    $ go get github.com/edwarnicke/pkgcloud/

## Client Usage

### Get all packages in a repo
```/bin/bash
pkgcloud all <user/repo>
```

### Get all packages with Custom Template

```/bin/bash
pkgcloud all <user/repo> -t <template>
```

The template is in [Go template format](https://golang.org/pkg/text/template/)

The following fields are available:

* {{.Name}} - The name of the package - Example: "jake"
* {{.CreatedAt}} - When the package was uploaded - Example: ""2017-03-13T02:49:29.000Z""
* {{.Epoch}} - The epoch of the package (if available) - Example: 0
* {{.Scope}} - The scope of the package (if available)- Example: null
* {{.Private}} - Whether or not the package is in a private repository - Example: false
* {{.UploaderName}} -  The name of the uploader for the package. - Example: "test_user"
* {{.Indexed}} - Whether or not this package has been indexed. - Example: false
* {{.RepositoryHTMLURL}} -  The HTML url of the repository. - Example: "/test_user/test_repo"
* {{.DownloadDetailsURL}} -   The url to get access log details for package downloads. - Example: "/api/v1/repos/test_user/test_repo/package/rpm/fedora/22/jake/x86_64/1.0/1.el6/stats/downloads/detail.json"
* {{.DownloadSeriesURL}} - The url to get time series data for package downloads. - Example: "/api/v1/repos/test_user/test_repo/package/rpm/fedora/22/jake/x86_64/1.0/1.el6/stats/downloads/series/daily.json"
* {{.DownloadCountURL}} - The url to get the total number of package downloads.  - Example: "/api/v1/repos/test_user/test_repo/package/rpm/fedora/22/jake/x86_64/1.0/1.el6/stats/downloads/count.json"
* {{.PromoteURL}} - The url for promoting this to another repository. - Example: "/api/v1/repos/test_user/test_repo/fedora/22/jake-1.0-1.el6.x86_64.rpm/promote.json"
* {{.DestroyURL}} -  The url for the HTTP DELETE request to destroy this package - Example: "/api/v1/repos/test_user/test_repo/fedora/22/jake-1.0-1.el6.x86_64.rpm"
* {{.Filename}} - The filename of the package.   Example: "jake-1.0-1.el6.x86_64.rpm"
* {{.DistroVersion}} - The distro_version for the package. - "fedora/22"
* {{.Version}} - The version of the package. - Example: "1.0"
* {{.Release}} - The release of the package (if available) - "1.el6"
* {{.Type}} - The type of package ("deb", "gem", or "rpm"). - Example: "rpm"
* {{.PackageURL}} - The API url for this package - Example: "/api/v1/repos/test_user/test_repo/package/rpm/fedora/22/jake/x86_64/1.0/1.el6.json"
* {{.PackageHTMLURL }} - The HTML url for this package - Example: "/test_user/test_repo/packages/fedora/22/jake-1.0-1.el6.x86_64.rpm"

In addition, some 'methods' are provided:

* {{.DaysOld}} - Number of days since the package has been uploaded.  Derived from {{.CreatedAt}}
* {{.Promote "user/repo"}} - Promote the package to the named repo.  Note: Does have side effects to packagecloud.io *unless* you use -d or --dry-run flags.

#### Example: Filter for only packages with {{.Release}} equal "release"

```bash
pkgcloud all fdio/1710 -t $'{{if eq .Release "release"}}{{.PackageHTMLURL}}\n{{end}}' 
```

which produces output:
```
/fdio/1710/packages/ubuntu/xenial/vpp-api-java_17.10-release_amd64.deb
/fdio/1710/packages/ubuntu/xenial/vpp-api-lua_17.10-release_amd64.deb
/fdio/1710/packages/ubuntu/xenial/vpp-dbg_17.10-release_amd64.deb
/fdio/1710/packages/ubuntu/xenial/vpp_17.10-release_amd64.deb
/fdio/1710/packages/ubuntu/xenial/vpp-dev_17.10-release_amd64.deb
/fdio/1710/packages/ubuntu/xenial/vpp-lib_17.10-release_amd64.deb
/fdio/1710/packages/ubuntu/xenial/vpp-plugins_17.10-release_amd64.deb
/fdio/1710/packages/ubuntu/xenial/vpp-api-python_17.10-release_amd64.deb
/fdio/1710/packages/el/7/vpp-lib-17.10-release.x86_64.rpm
/fdio/1710/packages/el/7/vpp-devel-17.10-release.x86_64.rpm
/fdio/1710/packages/el/7/vpp-plugins-17.10-release.x86_64.rpm
/fdio/1710/packages/el/7/vpp-api-lua-17.10-release.x86_64.rpm
/fdio/1710/packages/el/7/vpp-api-java-17.10-release.x86_64.rpm
/fdio/1710/packages/el/7/vpp-api-python-17.10-release.x86_64.rpm
/fdio/1710/packages/el/7/vpp-17.10-release.x86_64.rpm
```

Lets break down some things about the commandline.
First ```$'string\n'``` escapes the string so that newlines can be represented with ```\n```.
Second ```{{if eq .Release "release"}}...{{end}}``` is an if statement the body of which is only evaluated if ```Release == "release"```
Third  ```{{.PackageHTMLURL}}\n``` outputs the package html url for any packages that match the if statement

#### Example: Promote packages with {{.Release}} equal "release" to "fdio/staging"

```bash
pkgcloud all fdio/1804 -t $'{{if eq .Release "release"}}{{.Promote "fdio/staging" }}\n{{end}}' -d
```

```{{.Promote "fdio/staging" }}``` Promotes any packages that that passthe if statement to repo "fdio/staging".

```-d``` causes this to be a 'dry run'... meaning it doesn't actually promote, just tells you what it would do to promote.

#### Example: Filter for only packages older than 415 days

```bash
pkgcloud all fdio/1707 -t $'{{if gt .DaysOld 415}}{{.PackageHTMLURL}}: {{.DaysOld}}\n{{end}}'
```

Which outputs:
```
/fdio/1707/packages/ubuntu/xenial/vpp_17.07-rc1~b2_amd64.deb: 416
/fdio/1707/packages/ubuntu/xenial/vpp-plugins_17.07-rc1~b2_amd64.deb: 416
/fdio/1707/packages/ubuntu/xenial/vpp-api-python_17.07-rc1~b2_amd64.deb: 416
/fdio/1707/packages/ubuntu/xenial/vpp-dbg_17.07-rc1~b2_amd64.deb: 416
/fdio/1707/packages/ubuntu/xenial/vpp-dev_17.07-rc1~b2_amd64.deb: 416
/fdio/1707/packages/ubuntu/xenial/vpp-lib_17.07-rc1~b2_amd64.deb: 416
/fdio/1707/packages/ubuntu/xenial/vpp-api-lua_17.07-rc1~b2_amd64.deb: 416
/fdio/1707/packages/ubuntu/xenial/vpp-dpdk-dkms_17.05-vpp5_amd64.deb: 416
/fdio/1707/packages/el/7/vpp-dpdk-devel-17.05-vpp5.x86_64.rpm: 416
```

#### Example: Filter for only packages older than 415 days and promote them to repo "fdio/backup"
```bash
pkgcloud all fdio/1707 -t $'{{if gt .DaysOld 415}}{{.Promote "fdio/backup"}}\n{{end}}' -d
```
In this command, ```{{.Promote "fdio/backup"}}``` promotes the packages that pass the filter of ```{{if gt .DaysOld 415}}``` to the repo "fdio/backup".
Note the -d, which causes this to be a dry run.  If you really want to perform the promote, remove the -d


#### Example: Filter for only packages older than 475 days and delete them:
```bash
pkgcloud all fdio/backup -t $'{{if gt .DaysOld 475}}{{.Destroy}}\n{{end}}' -d
```

In this command, ```{{.Destroy}}``` deletes the packages that pass the filter of ```{{if gt .DaysOld 415}}```.
Note the -d, which causes this to be a dry run.  If you really want to perform the delete, remove the -d


### Pushing packages

```bash
pkgcloud push user/repo/distro/version/ filename
```

There are two optional flags for ```pkgcloud push```:
* -d or --dry-run: which will tell you what would be done for pushing the package, but will not in fact push it, or delete if used in conjunction with -f
* -f or --force: If and only if the package to-be-pushed already exists in packagecloud.io, delete it and then push.

# Acknowledgement

This is based on the [wonderful golang pkgcloud package provided by Mathias Lafeldt](https://github.com/mlafeldt/pkgcloud).


