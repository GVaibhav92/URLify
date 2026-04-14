# Build 
FROM golang:1.25-alpine AS builder

RUN apk add --no-cache git

WORKDIR /app

# Copy dependency files first — Docker layer cache means
# this layer only rebuilds when go.mod/go.sum change,
# not on every code change
COPY go.mod go.sum ./
RUN go mod download

# Copy source and build
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o urlify ./cmd/server/main.go

# Runtime stage
FROM alpine:3.19

RUN apk add --no-cache ca-certificates tzdata

WORKDIR /app

# Copy only the compiled binary from builder stage
COPY --from=builder /app/urlify .

EXPOSE 8080

CMD ["./urlify"]