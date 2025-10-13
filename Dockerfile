# Simple Dockerfile for the Go API
FROM golang:1.20-alpine AS build
WORKDIR /app
COPY . .
RUN go mod download
RUN go build -o /delivery-app ./cmd

FROM alpine:3.18
COPY --from=build /delivery-app /delivery-app
EXPOSE 8080
CMD ["/delivery-app"]
