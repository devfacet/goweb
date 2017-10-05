# goweb

[![Release][release-image]][release-url] [![Build Status][build-image]][build-url] [![Coverage][coverage-image]][coverage-url] [![GoDoc][doc-image]][doc-url]

A Go library for building tiny web applications such as dashboards, SPAs, etc.

## Features

- Standard library compatible 
- Listening multiple ports
- Upload file handling
- Templating
- Content detection
- No external dependency
- High code coverage

## Installation

```bash
go get github.com/devfacet/goweb
```

## Usage

### A basic app

See [basic](examples/basic/main.go) for full code.

```go
// Init the server
web := server.New(server.Options{
  ID:      "web",
  Address: "localhost:3000",
})

// Pages
pages := []page.Options{
  page.Options{
    URLPath:    "/",
    FilePath:   "/templates/index.html",
    FileSystem: &fs,
    TemplateData: PageData{
      Title:       pageTitle,
      Description: pageDesc,
      Content:     "Hello world",
    },
  },
  page.Options{
    URLPath:      "foo",
    Content:      "{{.Body}}",
    TemplateData: struct{ Body string }{Body: "foo"},
  },
  page.Options{
    URLPath:      "bar/",
    Content:      "{{.Body}}",
    MatchAll:     true,
    TemplateData: struct{ Body string }{Body: "bar"},
  },
}

for _, v := range pages {
  p, err := page.New(v)
  if err != nil {
    log.Logger.Fatal(err)
  }
  if err := web.AddPage(p); err != nil {
    log.Logger.Fatal(err)
  }
}

// Listen
if err := web.Listen(); err != nil {
  log.Logger.Fatal(err)
}
```

```bash
cd examples/basic/
go build .
./basic
```

## Build

```bash
go build .
```

## Test

```bash
./test.sh
```

## Release

```bash
git add CHANGELOG.md # update CHANGELOG.md
./release.sh v1.0.0  # replace "v1.0.0" with new version

git ls-remote --tags # check the new tag
```

## Contributing

- Code contributions must be through pull requests
- Run tests, linting and formatting before a pull request (`test.sh`)
- Pull requests can not be merged without being reviewed
- Use "Issues" for bug reports, feature requests and discussions
- Do not refactor existing code without a discussion
- Do not add a new third party dependency without a discussion
- Use semantic versioning and git tags for versioning

## License

Licensed under The MIT License (MIT)  
For the full copyright and license information, please view the LICENSE.txt file.


[release-url]: https://github.com/devfacet/goweb/releases/latest
[release-image]: https://img.shields.io/github/release/devfacet/goweb.svg

[build-url]: https://travis-ci.org/devfacet/goweb
[build-image]: https://travis-ci.org/devfacet/goweb.svg?branch=master

[coverage-url]: https://coveralls.io/github/devfacet/goweb?branch=master
[coverage-image]: https://coveralls.io/repos/devfacet/goweb/badge.svg?branch=master&service=github

[doc-url]: https://godoc.org/github.com/devfacet/goweb
[doc-image]: https://godoc.org/github.com/devfacet/goweb?status.svg
