FROM golang:1.25.4-alpine3.22 as builder
COPY . .
RUN go build -o downdetectordetectorg .

FROM alpine:3.22
COPY --from=builder /go/downdetectordetectorg .

ENTRYPOINT ["./downdetectordetectorg"]
EXPOSE 8080
