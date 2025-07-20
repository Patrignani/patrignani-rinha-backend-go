# Stage 1: Build
FROM golang:1.24.1-alpine AS builder

WORKDIR /app

RUN apk add --no-cache git

COPY . .

RUN go build -ldflags="-s -w" -o main

# Stage 2: Runtime
FROM gcr.io/distroless/static:nonroot

WORKDIR /root/

COPY --from=builder /app/main .

ENV PORT=8888
EXPOSE 8888
USER nonroot

# Entrypoint
CMD ["./main"]
