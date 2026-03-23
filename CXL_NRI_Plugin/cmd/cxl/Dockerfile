ARG GO_VERSION=1.25

FROM golang:${GO_VERSION}-bookworm AS builder

WORKDIR /go/builder

# Fetch go dependencies in a separate layer for caching
COPY go.mod go.sum .
RUN --mount=type=cache,target=/go/pkg/mod/ go mod download

# Build nri-cxl plugin
COPY . .

RUN --mount=type=cache,target=/go/pkg/mod/ \
    make GO_BUILD="CGO_ENABLED=0 go build" \
         GO_BUILD_FLAGS="-ldflags '-extldflags=-static'" \
         clean build/bin/cxl

FROM gcr.io/distroless/static

COPY --from=builder /go/builder/build/bin/cxl /bin/nri-cxl

ENTRYPOINT ["/bin/nri-cxl", "-idx", "95"]
