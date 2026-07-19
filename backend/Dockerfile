# --- Stage 1: Build Frontend (SPA) ---
FROM oven/bun:alpine AS frontend-builder
WORKDIR /app/frontend

# Install dependencies
COPY frontend/package.json frontend/bun.lock ./
RUN bun install

# Build the frontend as static SPA
COPY frontend/ ./
RUN bun run build

# --- Stage 2: Build Backend ---
FROM golang:1.25-alpine AS backend-builder
WORKDIR /app

# Install deps
COPY backend/go.mod backend/go.sum ./
RUN go mod download

# Copy source code
COPY backend/ ./

# Build backend binary
RUN CGO_ENABLED=0 GOOS=linux go build -o pos-server ./cmd/app

# --- Stage 3: Final Image (minimal, no Node.js) ---
FROM alpine:3.21
WORKDIR /app

# Add ca-certificates for HTTPS requests
RUN apk add --no-cache ca-certificates

# Copy backend binary
COPY --from=backend-builder /app/pos-server ./pos-server

# Copy frontend static files
COPY --from=frontend-builder /app/frontend/dist ./frontend/dist

# Copy swagger docs
COPY --from=backend-builder /app/docs ./docs

# Expose only backend port
EXPOSE 8080

CMD ["./pos-server"]
