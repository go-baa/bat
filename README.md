# [Bat](http://go-baa.github.io/bat) [![GoDoc](http://img.shields.io/badge/go-documentation-blue.svg?style=flat-square)](http://godoc.org/github.com/go-baa/bat) [![License](http://img.shields.io/badge/license-mit-blue.svg?style=flat-square)](https://raw.githubusercontent.com/go-baa/bat/master/LICENSE) [![Build Status](http://img.shields.io/travis/go-baa/bat.svg?style=flat-square)](https://travis-ci.org/go-baa/bat)

run the app which can hot compile

``copy from bee``


## Getting Started


Install:

```
go get -u github.com/go-baa/bat
```

Command:

```
bat run
```

```
bat run -t .ini -t .go -t .html -e logs
```

> bat default open godeps option, every build will resave vendor

disable godeps check:

```
bat run -godeps=false
```
