FROM golang:1.26-alpine AS build

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -trimpath -o /im-server ./cmd/server

FROM alpine:3.22

WORKDIR /app
COPY --from=build /im-server ./im-server
COPY web ./web

USER 65532:65532
EXPOSE 8080 8081
ENTRYPOINT ["./im-server"]
