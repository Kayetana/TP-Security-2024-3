FROM golang:1.21-alpine AS build

WORKDIR /source
COPY . .

ENV CGO_ENABLED 0
ENV GOOS linux

RUN go build -o bin/proxy ./cmd/proxy/main.go

FROM ubuntu:20.04 AS release

COPY --from=build /source/bin/proxy /proxy

EXPOSE 8080

CMD ["/proxy"]
