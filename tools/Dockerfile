FROM golang:1.24.4-alpine3.21 AS builder

ADD . /go/src/tcli
WORKDIR /go/src/tcli

# build the source
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
go build -o bin/tcli \
-ldflags "-s -w" cmd/main.go

# use a minimal alpine image
FROM alpine:3.21

# make tcli available in path
COPY --from=builder /go/src/tcli/bin/tcli /usr/local/bin/tcli
COPY --from=builder /go/src/tcli/tools/config.yaml /home/tcli/.tcli/config.yaml
COPY --from=builder /go/src/tcli/tools/modules.yaml /home/tcli/.tcli/modules.yaml
COPY --from=builder /go/src/tcli/tools/data /home/tcli/.tcli/data

# Create a lower privileged user account to run tcli
RUN adduser --disabled-password --gecos "" --home "/home/tcli" \
    --shell "/sbin/nologin" --no-create-home --uid 1001 tcli && \
    chown -R tcli:tcli /home/tcli

USER 1001

# run the binary
ENTRYPOINT ["tcli"]
