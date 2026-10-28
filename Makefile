.PHONY: check test example package bootstrap acquire-client stage-release generate check-generator test-generator build-tools check-package

check:
	./check-all.sh

test:
	GOTOOLCHAIN=local GOWORK=off go test ./...
	GOTOOLCHAIN=local GOWORK=off go test -C tools/cdn ./...

example:
	bin/wow-teal build examples/WowTealExample/src --output examples/WowTealExample/build

package:
	bin/wow-teal package

bootstrap: acquire-client
	bin/wow-types bootstrap

acquire-client:
	bin/wow-types acquire-client

stage-release:
	bin/wow-types stage-client
	bin/wow-types stage-release

generate:
	bin/wow-types generate

check-generator:
	bin/wow-types check-generator

test-generator:
	GOTOOLCHAIN=local GOWORK=off go test ./internal/wowapi ./internal/release
	GOTOOLCHAIN=local GOWORK=off go test -C tools/cdn ./...

build-tools:
	mkdir -p build/bin
	GOTOOLCHAIN=local GOWORK=off go build -mod=readonly -o build/bin/wow-teal ./tools/wow-teal
	GOTOOLCHAIN=local GOWORK=off go build -mod=readonly -o build/bin/wow-types ./tools/wow-types

check-package:
	./scripts/check-consumer-archive.sh
