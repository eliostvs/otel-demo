ifeq ($(origin .RECIPEPREFIX), undefined)
  $(error This Make does not support .RECIPEPREFIX. Please use GNU Make 4.0 or later)
endif

.DEFAULT_GOAL  = help
.DELETE_ON_ERROR:
.ONESHELL:
.SHELLFLAGS    := -eu -o pipefail -c
.SILENT:
MAKEFLAGS      += --no-builtin-rules
MAKEFLAGS      += --warn-undefined-variables
SHELL          = bash

DEV_MARKER     = .__dev
OSFLAG         ?=
args           ?=
pkg            ?=./...

ifeq ($(OS),Windows_NT)
	OSFLAG = "windows"
else
	OSFLAG = $(shell uname -s)
endif

## help: print this help message
.PHONY: help
help:
	echo 'Usage:'
	sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' | sed -e 's/^/ /' | sort

## clean: delete development environment
.PHONY: clean
clean:
	rm $(DEV_MARKER) 2> /dev/null || true

$(DEV_MARKER):
	go mod download
	touch $(DEV_MARKER)

## dev: prepare development environment
.PHONY: dev
dev: $(DEV_MARKER)

## deps/outdated: list outdated dependencies
.PHONY: deps/outdated
deps/outdated:
	go list -f "{{if and (not .Main) (not .Indirect)}} {{if .Update}} {{.Update}} {{end}} {{end}}" -m -u all 2> /dev/null | awk NF

## deps/tidy: remove unused and check hash of the dependencies
deps/tidy:
	go mod tidy
	go mod verify
	go mod download

## deps/upgrade [pkg]: upgrade dependencies
.PHONY: deps/upgrade
deps/upgrade: deps/tidy
	go get -u $(pkg)

.PHONY: deploy/up
deploy/up:
	docker-compose up --remove-orphans --build --detach

.PHONY: deploy/down
deploy/down:
	docker-compose down

.PHONY: run/digit
run/digit:
	go run cmd/digit/main.go --port 5001

.PHONY: run/lower
run/lower:
	DIGIT_URL=http://localhost:5001/ go run cmd/lower/main.go --port 5002

.PHONY: run/upper
run/upper:
	go run cmd/upper/main.go --port 5003

.PHONY: run/special
run/special:
	go run cmd/special/main.go --port 5004

.PHONY: run/generator
run/generator:
	DIGIT_URL=http://localhost:5001/ \
	LOWER_URL=http://localhost:5002/ \
	UPPER_URL=http://localhost:5003/ \
	SPECIAL_URL=http://localhost:5004/ \
	go run cmd/generator/main.go --port 5000

.PHONY: run/load
run/load:
	GENERATOR_URL=http://localhost:5000 go run cmd/load/main.go

.PHONY: run/all
run/all: run/digit run/lower run/upper run/special run/generator run/load

.PHONY: open/uptrace
open/uptrace:
	browse http://localhost:14318

.PHONY: open/grafana
open/grafana:
	browse http://localhost:3000

.PHONY: open/prometheus
open/prometheus:
	browse http://localhost:9090