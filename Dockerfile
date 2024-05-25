FROM golang:1.22.3-alpine as build
WORKDIR /app
COPY . .
RUN go mod download
RUN go build -o app-runnable ./cmd/*.go

FROM alpine:3.18
COPY --from=build /app/app-runnable ./app
COPY --from=build /app/app.yaml ./app.yaml
EXPOSE 8080
ENTRYPOINT ["./app"]