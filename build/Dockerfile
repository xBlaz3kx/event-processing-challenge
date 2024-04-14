FROM golang:1.21.1 as build

WORKDIR /app/src

ENV GOCACHE=/root/.cache/go-build
ENV GOMODCACHE=/root/.cache/go-build
ENV GO111MODULE=on
ENV CGO_ENABLED=0
ENV GOOS=linux

# Leverage build cache by copying go.mod and go.sum first
COPY go.mod go.sum ./
RUN --mount=type=cache,target=/root/.cache/go-build go mod download
RUN go mod verify

# Copy the rest of the source code
COPY . .

RUN --mount=type=cache,target="/root/.cache/go-build" go build -o ../service .

FROM alpine as app

SHELL ["/bin/sh", "-l", "-c"]

# Add curl for health checks
RUN apk --update --no-cache add curl

# Copy the service binary
COPY --chown=$user:$group --from=build /app/service /usr/local/bin/service
RUN chmod +x /usr/local/bin/service

ARG user=app
ARG group=docker

RUN addgroup $group
RUN adduser \
    --disabled-password \
    --gecos "" \
    --ingroup $group \
    $user \
USER $user

WORKDIR /usr/local/service

ENTRYPOINT ["/usr/local/bin/service"]