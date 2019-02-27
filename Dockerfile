FROM golang:1.12-alpine3.9 as build
RUN mkdir /src
WORKDIR /src
RUN apk --update add \
  git ca-certificates build-base
COPY . .
RUN go mod download
RUN go build ./cmd/wgo

FROM alpine:3.9
RUN mkdir /app
COPY --from=build /src/wgo /app/

ENTRYPOINT [ "/app/wgo" ]
