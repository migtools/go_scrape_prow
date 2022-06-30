LANG_DIR ?= go

MK_PATH := $(abspath $(lastword $(MAKEFILE_LIST)))
PROJECT_DIR := $(dir $(MK_PATH))
include $(PROJECT_DIR)/Makefile_common.mk

.PHONY: tests unit-tests

## tests: runs all available tests
tests: \
	unit-tests \
	golangci-lint

## unit-tests: runs the go unit-tests
unit-tests: get-common-ci-tools $(LANG_DIR)-dev-env
	. ${PYTHON_VENV}/bin/activate && \
	./scripts/run-unit-tests

# Just a help to the target specified in the Makefile_common.mk
# Help is not working from included Makefiles
## get-common-ci-tools: syncs the common ci upstream repoitory
## go-dev-env: installs go virtual env 
## go-clean-dev-env: cleans go dev environment 
## go-version: prints the go version being used by the Makefile
## golangci-lint: checks go lints 
## pre-commit: runs pre-commit rules set for projects 
