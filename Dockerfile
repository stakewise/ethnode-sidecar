FROM golang:1.17-alpine as build
WORKDIR /app
COPY . .
RUN go mod tidy && GOOS=linux go build -o sidecar main.go

FROM scratch
WORKDIR /app
COPY --from=build /app/config.yml .
COPY --from=build /app/sidecar .
ENTRYPOINT ["/app/sidecar"]
