#Common CI tools repo
SHELL = bash
COMMON_MK_DIR  := .konveyor
COMMON_GIT_DIR := .konveyor_repo

COMMON_CI_TOOLS_REPO := https://github.com/konveyor/common-ci-config

MK_PATH := $(abspath $(lastword $(MAKEFILE_LIST)))
PROJECT_DIR := $(dir $(MK_PATH))
COMMON_CI_TOOLS_REPO_DIR := $(PROJECT_DIR)/$(COMMON_GIT_DIR)


$(COMMON_MK_DIR):
		pushd "$(PROJECT_DIR)"; \
		ln -s "$(COMMON_CI_TOOLS_REPO_DIR)/makefile/.konveyor" "./$(COMMON_MK_DIR)"; \
		popd; \

.PHONY: get-common-ci-tools
get-common-ci-tools: $(COMMON_MK_DIR)
	if [ -d  "$(COMMON_CI_TOOLS_REPO_DIR)" ]; then \
		echo "Common CI tools exists"; \
		pushd "$(COMMON_CI_TOOLS_REPO_DIR)"; \
		git pull --rebase; \
		popd; \
	else \
		git clone $(COMMON_CI_TOOLS_REPO) "$(COMMON_CI_TOOLS_REPO_DIR)"; \
	fi

ifneq ("$(wildcard ./$(COMMON_GIT_DIR)/makefile/Makefile)","")
    include ./$(COMMON_GIT_DIR)/makefile/Makefile
endif

