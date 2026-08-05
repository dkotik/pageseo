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
generate:
	@clear
	@output=$$(go generate ./... && go test . -update) || echo "$$output"
	@date +"[ %T ]"
	@echo "Popular pages had 44 failures, new count is:" `cat testdata/popular.golden | grep "        --- FAIL: TestPopularPages" | wc -l`
build:
	goreleaser release --snapshot --rm-dist
install:
	cd ./cmd/pageseo && go build -trimpath -o ~/.local/bin/pageseo
	chmod +x ~/.local/bin/pageseo
