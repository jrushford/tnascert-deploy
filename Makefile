#
# Copyright (C) 2025 by John J. Rushford jrushford@apache.org
#
# This program is free software: you can redistribute it and/or modify
# it under the terms of the GNU General Public License as published by
# the Free Software Foundation, either version 3 of the License, or
# (at your option) any later version.
#
# This program is distributed in the hope that it will be useful,
# but WITHOUT ANY WARRANTY; without even the implied warranty of
# MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
# GNU General Public License for more details.
#
# You should have received a copy of the GNU General Public License
# along with this program.  If not, see <http://www.gnu.org/licenses/>.
#

.DEFAULT_GOAL := build

.PHONY: all
	
all:	fmt vet test build

fmt:
	go fmt ./...

vet: fmt
	go vet ./...

test: vet
	go test -cover ./...

build: test
	go get
	go build

run-tests:
	@echo "\nRunning main package tests"
	@go test -cover
	@echo "\nRunning config package tests"
	@(cd config && go test -cover)
	@echo "\nRunning clients package tests"
	@(cd clients && go test -cover )
	@echo "\nRunning clients/restapi package tests"
	@(cd clients/restapi && go test -cover )
	@echo "\nRunning clients/wsapi package tests"
	@(cd clients/wsapi && go test -cover )

clean: tnascert-deploy
	rm tnascert-deploy
