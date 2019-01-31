FROM golang:1.11-alpine3.8 as build
RUN mkdir /src
WORKDIR /src
RUN apk --update add \
  git ca-certificates build-base
COPY . .
RUN go mod download
RUN go build ./cmd/wgo

FROM alpine:3.8
RUN mkdir /app
COPY --from=build /src/wgo /app/

ENTRYPOINT [ "/app/wgo" ]
