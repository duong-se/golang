install:
	@GO111MODULE=on go mod vendor
sqlc:
	@docker run --rm -v $(pwd):/src -w /src sqlc/sqlc generate
dbmate:
	@docker run --rm  -v $(pwd):/src -w /src -it --network=host ghcr.io/amacneil/dbmate
