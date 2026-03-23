FROM golang:1.25-alpine AS build
WORKDIR /src
COPY go.mod ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/central ./main.go

FROM gcr.io/distroless/base-debian12
COPY --from=build /out/central /central
ENV PORT=8080
EXPOSE 8080
ENTRYPOINT ["/central"]
