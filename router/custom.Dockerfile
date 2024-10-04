FROM segment/chamber:2.12.0 AS chamber
FROM golang:1.21 AS builder

WORKDIR /app/

# Copy only the files required for go mod download
COPY ./go.* .

# Download dependencies
RUN go mod download

# Copy the rest of the files
COPY . .

# Run tests
RUN go test -v ./...

# Build router
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags "-extldflags -static -X github.com/wundergraph/cosmo/router/core.Version=${VERSION}" -a -o router cmd/cbs-sports/main.go

FROM gcr.io/distroless/static:latest

COPY --from=chamber /chamber /bin/chamber
COPY --from=builder /app/router /router

ENTRYPOINT ["/router"]

EXPOSE 3002
