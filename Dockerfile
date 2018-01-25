FROM golang:1.9-alpine as builder
RUN mkdir -p /go/src/github.com/blockloop/pbpaste
ADD . /go/src/github.com/blockloop/pbpaste
RUN go build -o /go/bin/pbpaste github.com/blockloop/pbpaste

FROM alpine
COPY --from=builder /go/bin/pbpaste /pbpaste
RUN apk add --update --no-cache ca-certificates
ENTRYPOINT ["/pbpaste"]
EXPOSE 3000
