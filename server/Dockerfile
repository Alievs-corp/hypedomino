FROM golang:1.27.3-bookworm AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/ ./cmd/server ./cmd/migrate

FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=build /out/server /out/migrate /
USER 65532:65532
EXPOSE 8080
ENTRYPOINT ["/server"]
