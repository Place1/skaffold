FROM golang:1.10.1-alpine3.7 as builder
COPY main.go .
RUN go build -o /app main.go

FROM alpine:3.7
ARG message
RUN test -n "${message}"
RUN echo "the message is ${message}"
COPY --from=builder /app .
CMD ["./app"]
