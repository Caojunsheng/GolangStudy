###1. golang UT coverage
When you want to scan all go files in a package, and want to calculate the ut coverage of this package

**Installation**

Get the necessary packages and its dependencies:

`$ go get github.com/axw/gocov/gocov`

`$ go get -u gopkg.in/matm/v1/gocov-html`


**Usage**

`$ gocov test ./... | gocov-html > result.html`

**Reference**

https://github.com/axw/gocov

https://github.com/matm/gocov-html
