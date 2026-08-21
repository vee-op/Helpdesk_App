FROM golang:1.26.6-alpine3.24 AS builder

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN go build -v -o /usr/local/bin/mygoapp .

FROM alpine:3.23.5

COPY --from=builder /usr/local/bin/mygoapp /usr/local/bin/mygoapp

EXPOSE 5001

CMD ["mygoapp"]