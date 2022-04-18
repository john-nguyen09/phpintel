#!/bin/bash

VERSION=0.0.12

go build -ldflags "-s -w -X main.version=$VERSION" -o dist/
