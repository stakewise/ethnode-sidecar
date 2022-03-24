FROM golang:1.18-alpine as build
WORKDIR /app
COPY . .
RUN go mod tidy && CGO_ENABLED=0 GOOS=linux go build -o sidecar main.go

FROM scratch
USER 1000
WORKDIR /app
COPY --from=build /app/config.yml .
COPY --from=build /app/sidecar .
ENTRYPOINT ["/app/sidecar"]
