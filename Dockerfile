# syntax=docker/dockerfile:1

############################
# Build stage
############################
FROM golang:1.22 AS builder
WORKDIR /workspace

COPY go.mod go.sum ./
RUN go mod download

COPY . .

# CGO 비활성화 + 정적 링크 빌드
ENV CGO_ENABLED=0 GOOS=linux GOARCH=amd64
RUN go build -tags 'sqlite_omit_load_extension' -o server .

############################
# Runtime stage
############################
FROM gcr.io/distroless/static-debian11

WORKDIR /
COPY --from=builder /workspace/server /server

EXPOSE 8080
ENTRYPOINT ["/server"]
