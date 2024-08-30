## v0.6.0

### Dependency Updates:

* Upgraded gophercloud to v2 (**major**)
* Updated various dependencies to their latest versions, including:
    * [github.com/sapcc/go-api-declarations](https://github.com/sapcc/go-api-declarations)
    * [github.com/sapcc/gophercloud-sapcc](https://github.com/sapcc/gophercloud-sapcc)
    * [github.com/gophercloud/gophercloud](https://github.com/gophercloud/gophercloud)
    * [github.com/gophercloud/utils](https://github.com/gophercloud/utils)
    * [github.com/spf13/viper](https://github.com/spf13/viper)
    * gopkg.in/cheggaaa/pb.v1 (**major**)
    * golangci/golangci-lint-action
    * github/codeql-action
    * actions/checkout
    * actions/setup-go
    * actions/dependency-review-action
    * alpine Docker image
    * golang Docker image

### Other Changes

* Adjusted to recent go-makefile-maker changes
* Enabled vendoring in Makefile.maker.yaml
* Fixed gocritic emptyStringTest
* Replaced password command with go-bits implementation