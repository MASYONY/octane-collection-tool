FROM golang:1.22-alpine AS builder
WORKDIR /src
COPY go.mod ./
COPY cmd ./cmd
COPY internal ./internal
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/octane-collection-tool ./cmd/octane-collection-tool

FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=builder /out/octane-collection-tool /usr/local/bin/octane-collection-tool
ENTRYPOINT ["/usr/local/bin/octane-collection-tool"]
