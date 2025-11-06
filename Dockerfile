FROM golang:1.25-alpine AS builder

# CG0_ENABLED to enable static links for libraries 
ENV CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64

WORKDIR /build
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o /app .

# Final lightweight stage
FROM alpine:3.21 AS final
COPY --from=builder /app /bin/app
EXPOSE 3000
CMD ["bin/app"]