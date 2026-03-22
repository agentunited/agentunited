FROM golang:1.25-alpine AS builder
WORKDIR /src
COPY go.mod go.sum* ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/billing-webhook ./...

FROM gcr.io/distroless/static-debian12
COPY --from=builder /out/billing-webhook /billing-webhook
EXPOSE 8080
ENTRYPOINT ["/billing-webhook"]
