LANG_DIR ?= go

MK_PATH := $(abspath $(lastword $(MAKEFILE_LIST)))
PROJECT_DIR := $(dir $(MK_PATH))
include $(PROJECT_DIR)/Makefile_common.mk

.PHONY: tests unit-tests

tests: \
	unit-tests \
	golangci-lint

unit-tests: get-common-ci-tools $(LANG_DIR)-dev-env
	. ${PYTHON_VENV}/bin/activate && \
	./scripts/run-unit-tests
