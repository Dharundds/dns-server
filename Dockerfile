FROM oven/bun:latest AS bun-builder

WORKDIR /app

COPY dns-server/frontend/package.json .

RUN bun install

COPY dns-server/frontend/ .

RUN bun vite build


FROM golang:1.24-alpine AS go-builder

WORKDIR /app

COPY dns-server/go.mod .
COPY dns-server/go.sum .

RUN go mod download

COPY dns-server/cmd/ ./cmd/
COPY dns-server/internal/ ./internal/

RUN go build -o dns-server cmd/dns/main.go

FROM alpine:latest 

WORKDIR /dns-server

COPY --from=go-builder /app/dns-server .
COPY --from=bun-builder /app/dist ./dist

EXPOSE 3002 8080 53

CMD ["./dns-server"]