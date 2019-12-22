FROM golang:latest AS builder
WORKDIR /go/src/github.com/telecom-tower/quote-of-the-day
COPY main.go go.mod go.sum ./
RUN go get .
RUN go build -o quote-of-the-day .

FROM alpine:latest  
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /go/src/github.com/telecom-tower/quote-of-the-day/quote-of-the-day quote-of-the-day
CMD ["./quote-of-the-day"] 