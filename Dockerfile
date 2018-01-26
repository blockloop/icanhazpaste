FROM golang:1.9-alpine as builder
RUN mkdir -p /go/src/github.com/blockloop/pbpaste
ADD . /go/src/github.com/blockloop/pbpaste
RUN go build -o /go/bin/pbpaste github.com/blockloop/pbpaste

FROM alpine
RUN apk add --update --no-cache ca-certificates
ENTRYPOINT ["/pbpaste"]
ADD CHECKS /app/CHECKS
ADD form.html /form.html
ADD styles.css /styles.css
COPY --from=builder /go/bin/pbpaste /pbpaste
EXPOSE 3000
