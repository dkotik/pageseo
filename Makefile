-include .env
export

default:
	@clear
	@output=$$(go test ./...) || echo "$$output"
	@date +"[ %T ]"
short:
	@clear
	@output=$$(go test -short ./...) || echo "$$output" | grep -Ev "^(ok|\\?)"
	@date +"[ %T ]"
build:
	goreleaser release --snapshot --rm-dist
install:
	cd ./cmd/pageseo && go build -trimpath -o ~/.local/bin/pageseo
	chmod +x ~/.local/bin/pageseo
