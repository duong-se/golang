FROM golang:1.25.11

RUN apt-get update && apt-get install -y \
    git curl vim ca-certificates \
    gnupg \
    && rm -rf /var/lib/apt/lists/*

RUN curl -fsSL -o /usr/local/bin/dbmate https://github.com/amacneil/dbmate/releases/latest/download/dbmate-linux-amd64
RUN chmod +x /usr/local/bin/dbmate

ENV GO111MODULE=on

WORKDIR /go/src/github.com/duong-se/golang
