FROM golang:1.23 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 go build -o content-prep .

FROM scratch

COPY --from=builder /app/content-prep /usr/bin/content-prep

ENTRYPOINT ["/usr/bin/content-prep"]