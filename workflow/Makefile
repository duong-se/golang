install:
	@go install github.com/pressly/goose/v3/cmd/goose@latest
	@GO111MODULE=on go mod vendor
sqlc:
	@docker run --rm -v $(pwd):/src -w /src sqlc/sqlc generate
dbmate:
	@docker run --rm  -v $(pwd):/src -w /src -it --network=host ghcr.io/amacneil/dbmate
