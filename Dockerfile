FROM golang:alpine AS builder

ENV GIN_MODE=release
ENV PORT=3004

WORKDIR /go/src/finlit

COPY . .

RUN go mod download
RUN go mod verify
RUN go build -o finlit


FROM alpine:3.8
RUN apk add --no-cache ca-certificates
COPY --from=builder /go/src/finlit .
ENTRYPOINT ["./finlit"]