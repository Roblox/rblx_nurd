#!/bin/bash

set -eo pipefail

export PATH=$PATH:/usr/local/go/bin
export PATH=$PATH:/usr/local/bin
export GO_VERSION=1.14.3

main() {
	if [ ! -z "$CIRCLECI" ]; then
		# Remove default golang (1.7.3) and install a custom version (1.14.3) of golang.
		# This is required for supporting go mod, and to be able to compile nurd.
		sudo rm -rf /usr/local/go

		# Install golang 1.14.3
		curl -L -o go${GO_VERSION}.linux-amd64.tar.gz https://dl.google.com/go/go${GO_VERSION}.linux-amd64.tar.gz
		sudo tar -C /usr/local -xzf go${GO_VERSION}.linux-amd64.tar.gz
		sudo chmod +x /usr/local/go
		rm -f go${GO_VERSION}.linux-amd64.tar.gz
	fi

	# Run tests
	trap 'docker-compose down' exit
	docker-compose up -d mssql
	go test -cover -count=1 -v ./...
}

main "$@"