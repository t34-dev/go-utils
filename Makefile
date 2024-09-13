APP_NAME := go-utils
APP_REPOSITORY := gitlab.com/zakon47
%:
	@:
MAKEFLAGS += --no-print-directory
export GOPRIVATE=$(APP_REPOSITORY)/*
export GOPRIVATE=github.com/t34-dev/*
export COMPOSE_PROJECT_NAME
# ============================== Environments
ENV ?= local
FOLDER ?= /root
VERSION ?=
# ============================== Paths
APP_EXT := $(if $(filter Windows_NT,$(OS)),.exe)
BIN_DIR := $(CURDIR)/.bin
DEVOPS_DIR := $(CURDIR)/.devops
TEMP_DIR := $(CURDIR)/.temp
ENV_FILE := $(CURDIR)/.env
SECRET_FILE := $(CURDIR)/.secrets
CONFIG_DIR := $(CURDIR)/configs
# ============================== Includes
include make/get-started.mk
include make/tag.mk
