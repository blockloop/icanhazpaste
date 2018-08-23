FROM golang:1.9-alpine as builder
RUN mkdir -p /go/src/github.com/blockloop/icanhazpaste
ADD . /go/src/github.com/blockloop/icanhazpaste
RUN go build -o /go/bin/icanhazpaste github.com/blockloop/icanhazpaste

FROM alpine
RUN apk add --update --no-cache ca-certificates
ENTRYPOINT ["/icanhazpaste"]
ADD CHECKS /app/CHECKS
ADD form.html /form.html
ADD styles.css /styles.css
COPY --from=builder /go/bin/icanhazpaste /icanhazpaste
EXPOSE 3000
