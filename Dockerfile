FROM golang:alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o main .

# Use distroless static image as the base image for the final stage
FROM gcr.io/distroless/static

COPY --from=builder /app/main /main

EXPOSE 8080

ENTRYPOINT ["/main"]
