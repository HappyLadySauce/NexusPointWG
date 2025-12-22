# Use bash if available, otherwise use default shell
ifeq ($(findstring Windows,$(OS)),)
SHELL := /bin/bash
endif

# Define COMMON_SELF_DIR if not already defined
# This gets the directory of the current Makefile (common.mk)
ifeq ($(origin COMMON_SELF_DIR),undefined)
COMMON_SELF_DIR := $(dir $(lastword $(MAKEFILE_LIST)))
endif

# Define ROOT_DIR if not already defined
# Go up two levels from scripts/make-rules/ to get project root
ifeq ($(origin ROOT_DIR),undefined)
# Use shell command to get absolute path of project root
# COMMON_SELF_DIR is scripts/make-rules/, go up two levels to get project root
ifeq ($(OS),Windows_NT)
# On Windows, use PowerShell to get parent of parent directory
ROOT_DIR := $(shell powershell -Command "(Get-Item '$(COMMON_SELF_DIR)').Parent.Parent.FullName")
else
# On Unix, use cd and pwd
ROOT_DIR := $(shell cd $(COMMON_SELF_DIR)/../.. && pwd)
endif
endif

# Define OUTPUT_DIR if not already defined
ifeq ($(origin OUTPUT_DIR),undefined)
OUTPUT_DIR := $(ROOT_DIR)/_output
# Create output directory if it doesn't exist
# Use if not exist for Windows, mkdir -p for Unix
ifeq ($(OS),Windows_NT)
$(shell if not exist "$(OUTPUT_DIR)" mkdir "$(OUTPUT_DIR)")
else
$(shell mkdir -p "$(OUTPUT_DIR)")
endif
endif

# Specify tools severity, include: BLOCKER_TOOLS, CRITICAL_TOOLS, TRIVIAL_TOOLS.
# Missing BLOCKER_TOOLS can cause the CI flow execution failed, i.e. `make all` failed.
# Missing CRITICAL_TOOLS can lead to some necessary operations failed. i.e. `make release` failed.
# TRIVIAL_TOOLS are Optional tools, missing these tool have no affect.
BLOCKER_TOOLS ?= gsemver golines go-junit-report golangci-lint addlicense goimports codegen
CRITICAL_TOOLS ?= swagger mockgen gotests git-chglog github-release coscmd go-mod-outdated protoc-gen-go cfssl go-gitlint
TRIVIAL_TOOLS ?= depth go-callvis gothanks richgo rts kube-score

COMMA := ,
SPACE :=
SPACE +=