FROM golang:1.26-alpine
WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/api ./cmd/api
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/seed ./cmd/seed

FROM gcr.io/distroless/static-debian13:nonroot
WORKDIR /app
COPY --from=0 /out/api /app/api
COPY --from=0 /out/seed /app/seed
USER nonroot:nonroot
EXPOSE 8080
ENTRYPOINT ["/app/api"]
