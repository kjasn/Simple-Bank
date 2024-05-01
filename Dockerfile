# build stage
FROM golang:1.22-alpine3.19 AS builder

ENV http_proxy "http://127.0.0.1:7890"
ENV https_proxy "http://127.0.0.1:7890"


WORKDIR /app
COPY . .
RUN go build -o main main.go

# run stage
FROM alpine:3.19
WORKDIR /app
COPY --from=builder /app/main .
COPY app.env .
EXPOSE 8080
CMD [ "/app/main" ]
