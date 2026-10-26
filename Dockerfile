FROM golang:1.27.3-alpine3.24 AS build

WORKDIR /src
COPY go.mod ./
COPY . .
RUN CGO_ENABLED=0 go build -trimpath -ldflags='-s -w' -o /url-checker ./cmd/server

FROM alpine:3.24.3
RUN adduser -D -H app
USER app
COPY --from=build /url-checker /url-checker
EXPOSE 8123
ENTRYPOINT ["/url-checker"]
