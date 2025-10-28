# syntax=docker/dockerfile:1  
# Build stage  
FROM golang:1.18 AS builder  
  
# Set the working directory inside the container  
WORKDIR /workspace  
  
# Copy go module files and download dependencies  
COPY go.mod go.sum ./  
RUN go mod download  
  
# Copy the rest of the source code  
COPY . .  
  
# Build the Go application for Linux  
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o server .  
  
# Final stage: use a minimal distroless base image  
FROM gcr.io/distroless/base-debian11  
  
# Copy the statically linked binary from the builder stage  
COPY --from=builder /workspace/server /server  
  
# Expose the port that the server listens on  
EXPOSE 8080  
  
# Command to run the binary  
ENTRYPOINT ["/server"]
