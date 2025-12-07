ARG release_image_tag
FROM golang:1.25.0 AS base
LABEL authors="guilherme passos"
COPY ./go.mod ./go.sum ./
RUN go mod download

FROM base AS build
#ENV CGO_ENABLED 1
RUN go env -w GOCACHE=/go-cache
COPY . .
RUN --mount=type=cache,target=/go-cache \
    go build -o /out/cmd ./cmd/

FROM gcr.io/distroless/base:${release_image_tag:-debug} AS release
COPY --from=build /out/cmd /
WORKDIR /go
COPY static/ /go/static
COPY templates/ /go/templates
COPY static/ /static
COPY templates/ /templates
ENTRYPOINT ["/cmd"]