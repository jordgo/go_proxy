FROM golang:alpine AS builder

LABEL stage=gobuilder

ENV CGO_ENABLED 0

ENV GOOS linux

#RUN apk update --no-cache && apk add --no-cache tzdata

WORKDIR /build

ADD go.mod .

ADD go.sum .

RUN go mod download

COPY . .

RUN go build -ldflags="-s -w" -tags timetzdata -o /app/go_proxy cmd/main.go

FROM alpine

RUN apk update --no-cache && apk add --no-cache ca-certificates

WORKDIR /app

COPY --from=builder /app/go_proxy /app/go_proxy

RUN mkdir /app/logs

CMD ["./go_proxy"]
