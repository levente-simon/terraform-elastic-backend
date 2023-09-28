FROM golang:1.21 AS builder
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -o terraform-backend .

FROM alpine:3.18
RUN apk --no-cache add ca-certificates
COPY --from=builder /src/terraform-backend /app/terraform-backend
ENTRYPOINT ["/app/terraform-backend"]
EXPOSE 8080 8443
