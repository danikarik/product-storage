FROM golang:1.15-alpine3.12 as builder
RUN mkdir -p /build
COPY . /build
WORKDIR /build
RUN CGO_ENABLED=0 GOOS=linux go build -a -o server ./cmd/server/main.go

FROM alpine:3.12
COPY --from=builder /build/server .

ENTRYPOINT ["./server"]
