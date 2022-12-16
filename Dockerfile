ARG GO_VERSION=1.19
FROM golang:${GO_VERSION}-alpine AS builder
ARG VERSION
ARG COMMIT

RUN apk add --update --no-cache ca-certificates make git curl
RUN mkdir -p /build

WORKDIR /build
COPY go.* /build/
COPY . /build
RUN go mod download

#RUN go build -ldflags="-X github.com/innovia/portal/version.version=${VERSION} -X github.com/innovia/portal/version.gitCommitID=${COMMIT}"
RUN make build
RUN cp bin/portal /usr/local/bin/
RUN chmod a+x /usr/local/bin/portal

FROM alpine

COPY --from=builder /usr/local/bin/portal /usr/local/bin/
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Avoid running as root or named user
USER 65534

ENTRYPOINT ["/usr/local/bin/portal"]
