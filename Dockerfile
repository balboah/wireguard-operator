FROM golang:1.15-alpine3.12 as build
RUN mkdir /src
WORKDIR /src
RUN apk --update add \
  git ca-certificates build-base
COPY . .
RUN go mod download
RUN go build ./cmd/wgo

FROM alpine:3.12
RUN mkdir /app
COPY --from=build /src/wgo /app/

ENTRYPOINT [ "/app/wgo" ]
