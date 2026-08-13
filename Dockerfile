FROM golang:1.26-alpine3.24 AS build
RUN mkdir /src
WORKDIR /src
RUN apk --update add \
  git ca-certificates build-base
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build ./cmd/wgo

FROM alpine:3.24
RUN mkdir /app
COPY --from=build /src/wgo /app/

ENTRYPOINT [ "/app/wgo" ]
