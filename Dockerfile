FROM golang:1.22.2-alpine as build-base

WORKDIR /app

COPY . .

RUN go mod download

RUN CGO_ENABLED=0 go test -v --tags=unit ./...

RUN go build -o ./out/server .

# ====================

FROM alpine:3.19.1

COPY --from=build-base /app/out/server /app/server

EXPOSE 8080

CMD ["/app/server"]
