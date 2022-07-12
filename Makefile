# Copyright (2022 -- present) Shahruk Hossain <shahruk10@gmail.com>
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#		 http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
# ==============================================================================

TOP := $(dir $(realpath $(firstword $(MAKEFILE_LIST))))
SHELL := /usr/bin/env bash

EXTRA_LDFLAGS := -linkmode external -extldflags '-static -static-libstdc++ -static-libgcc'

.PHONY: nix-build
nix-build:
	nix-shell --pure shell.nix --run "$(MAKE) build"

.PHONY: nix-test
nix-test:
	nix-shell --pure shell.nix --run "$(MAKE) test"

test:
	go test -race -cover ./...

.PHONY: build
build: sctk

.PHONY: clean
clean:
	rm -f $(TOP)/sctk $(TOP)/.built_sctk_module
	cd $(TOP)/SCTK && make clean

sctk: sctk-bins
	CGO_ENABLED=0 go build -o $(TOP)/sctk \
	-ldflags "-s -w $(EXTRA_LDFLAGS) $(LDFLAGSVERSION)" \
	./cmd/sctk

sctk-bins:
	if [[ ! -e $(TOP)/.built_sctk_module ]]; then \
		git submodule update --recursive \
		&& cd $(TOP)/SCTK && CFLAGS="-static" CPPFLAGS="-static" $(MAKE) clean config all install \
		&& touch $(TOP)/.built_sctk_module ; \
	fi

	cp $(TOP)/SCTK/bin/{sclite,sc_stats} $(TOP)/internal/sctk/embedded/bin

