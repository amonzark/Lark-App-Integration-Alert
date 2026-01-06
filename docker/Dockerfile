FROM golang:1.23-alpine AS builder

WORKDIR /app

COPY . .
RUN go mod download
RUN go build -o /katulampa-lark-app ./cmd

FROM alpine:latest

RUN apk --no-cache add curl
COPY --from=builder /katulampa-lark-app /bin/katulampa-lark-app
CMD ["/bin/katulampa-lark-app"]