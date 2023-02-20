Swagroller
==========

Swagroller is a command-line interface (CLI) tool that converts a YAML file to an OpenAPI specification document,
and generates a simple static web page that can be run anywhere, even just on a file system.

## Installation
You can Swagroller by using the following command:
```shell
$ go get github.com/scryner/swagroller
```

You can also clone the repository and build the binary from source:
```shell
git clone https://github.com/scryner/swagroller.git
cd swagroller
go build
```

## Usage
To build a static web page from a specification file in YAML, simply run the following command:
```shell
$ swagroller build <path-to-yaml-file>
```

You can also serve the contents to your web browser using the following command:
```shell
$ swagroller serv <path-to-yaml-file>
```
