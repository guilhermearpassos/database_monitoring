ARG release_image_tag
FROM golang:1.24.0 as base
LABEL authors="guilherme passos"
COPY ./go.mod ./go.sum ./
RUN go mod download

FROM base as build
#ENV CGO_ENABLED 1
RUN go env -w GOCACHE=/go-cache
COPY . .
RUN --mount=type=cache,target=/go-cache \
    go build -o /out/cmd ./cmd/

FROM gcr.io/distroless/base:${release_image_tag:-debug} AS release
COPY --from=build /out/cmd /
COPY static/ /go/static
COPY templates/ /go/templates
ENTRYPOINT ["/cmd"]