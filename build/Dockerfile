FROM golang:1.21-alpine AS build

WORKDIR /source
COPY . .

ENV CGO_ENABLED 0
ENV GOOS linux

RUN go build -o bin/proxy ./cmd/proxy/main.go

FROM alpine:3.17 AS release

RUN apk update && apk add --no-cache openssl ca-certificates

COPY --from=build /source/bin/proxy /proxy
COPY --from=build /source/certs /certs

EXPOSE 8080

CMD ["/proxy"]
