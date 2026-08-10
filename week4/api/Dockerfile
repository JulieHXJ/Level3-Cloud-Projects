# ---------- Build stage ----------
FROM --platform=$BUILDPLATFORM golang:1.25-alpine AS builder

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .

ARG TARGETOS
ARG TARGETARCH

RUN CGO_ENABLED=0 \
    GOOS=${TARGETOS:-linux} \
    GOARCH=${TARGETARCH:-amd64} \
    go build \
    -trimpath \
    -ldflags="-s -w" \
    -o /out/cloud4-api \
    .

# ---------- Runtime stage ----------
FROM gcr.io/distroless/static-debian12:nonroot

WORKDIR /app

COPY --from=builder /out/cloud4-api /app/cloud4-api

EXPOSE 8080

USER nonroot:nonroot

ENTRYPOINT ["/app/cloud4-api"]