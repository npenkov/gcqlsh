FROM golang AS builder

COPY . /app

WORKDIR /app

RUN go mod vendor
RUN CGO_ENABLED=0 go build -ldflags="-s -w -X main.version=$(git describe --tags --exact-match HEAD 2>/dev/null || git rev-parse --short HEAD 2>/dev/null || echo -n unknown | sed 's/^v//')" -o /gcqlsh cmd/gcqlsh.go


FROM alpine

COPY --from=builder /gcqlsh /bin/gcqlsh

ENTRYPOINT ["/bin/sh"]
