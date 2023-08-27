# syntax=docker/dockerfile:1

FROM golang:1.20

WORKDIR /app

COPY go.mod go.sum ./

COPY *.go **/*.go/ ./
RUN go mod download
COPY cmd/petrid/go.mod cmd/petrid/go.sum ./cmd/petrid/
COPY cmd/petrid/*.go cmd/petrid/**/*.go  ./cmd/petrid/

RUN go work init
RUN go work use .
RUN go work use cmd/petrid
RUN go work sync

WORKDIR /app/cmd/petrid
RUN go mod download


WORKDIR /app
EXPOSE 8081
RUN go run cmd/petrid/main.go

