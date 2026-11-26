ARG release_image_tag
FROM golang:1.27.3 AS base
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
ENTRYPOINT ["/cmd"]
