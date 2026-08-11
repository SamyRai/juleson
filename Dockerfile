FROM golang:1.26.4-alpine AS builder

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/juleson ./cmd/juleson && \
    CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/jsn ./cmd/juleson

FROM alpine:3.22

RUN addgroup -S juleson && adduser -S -G juleson juleson

WORKDIR /workspace
COPY --from=builder /out/juleson /usr/local/bin/juleson
COPY --from=builder /out/jsn /usr/local/bin/jsn

USER juleson
ENTRYPOINT ["juleson"]
