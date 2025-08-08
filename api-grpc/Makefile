DESTDIR ?= /var/local
MODBASE = multiple.api
NAME ?= ex
GIT_COMMIT:=$(shell git rev-parse --short HEAD)
BUILD_DATE:=$(shell date -u +'%Y-%m-%dT%H:%M:%SZ')
VERSION:=$(shell git describe --tags --abbrev=0 --match "svc-$(NAME)*" | tr -d '\n')

##
## pos-relay Makefile help
##
## Usage: make [target]
##
## Target includes: build
##
.PHONY: help
help: 			## Show this help
	@fgrep -h "##" $(MAKEFILE_LIST) | fgrep -v fgrep | sed -e 's/\\$$//' | sed -e 's/##//'


.PHONY: svc
svc:
	mkdir -p $(DESTDIR)/static
	go build -o $(DESTDIR)/apisvc \
	-ldflags=" \
	-X $(MODBASE)/utils/version.gitVersion=$(VERSION) \
	-X $(MODBASE)/utils/version.gitCommit=$(GIT_COMMIT) \
	-X $(MODBASE)/utils/version.buildDate=$(BUILD_DATE) \
	" cmd/server/$(NAME)/main.go
	cp static/swagger/* $(DESTDIR)/static/
	cp src/gen/api.swagger.yaml $(DESTDIR)/static/swagger.yaml