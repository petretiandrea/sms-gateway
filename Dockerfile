FROM golang:1.20.6-alpine as build
WORKDIR /app
COPY . .
RUN go mod download
RUN go build -o app-runnable ./cmd/main.go

FROM alpine:3.18
COPY --from=build /app/app-runnable ./app
EXPOSE 8080
ENTRYPOINT ["./app"]