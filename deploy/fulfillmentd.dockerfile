FROM golang:1.26-alpine AS builder
LABEL eluvio.container.author="John Heaton" \
      eluvio.code.repo="https://github.com/eluv-io/redeemable-fulfillment-sample"

WORKDIR /build

ARG GIT_COMMIT
ARG GIT_BRANCH=main

RUN apk --no-cache add git bash openssh-client

# Hope to get rid of this ---------------
# private qluvio modules are fetched over SSH from the build's agent (podman build --ssh default); no token
ENV GOPRIVATE=github.com/qluvio
RUN mkdir -p -m 0700 ~/.ssh && \
    ssh-keyscan github.com >> ~/.ssh/known_hosts && \
    git config --global url."ssh://git@github.com/".insteadOf "https://github.com/"
# ---------------------------------------

RUN --mount=type=ssh \
  git clone --branch ${GIT_BRANCH} https://github.com/eluv-io/redeemable-fulfillment-sample fulfillmentd && \
  cd fulfillmentd && \

  (cd version ; go generate) && \
  CGO_ENABLED=0 GOOS=linux go build -o /bin/fulfillmentd cmd/main.go && \
  ls -l /bin/fulfillmentd && \
  /bin/fulfillmentd

FROM alpine:latest
WORKDIR /
RUN \
  mkdir -p /default /logs /bin
COPY --from=builder /bin/fulfillmentd /bin/
COPY configs/fulfillmentd-test.toml /default/fulfillmentd.toml
CMD [ "/bin/fulfillmentd" ]
ENTRYPOINT [ "/bin/fulfillmentd", "--config", "/default/fulfillmentd.toml" ]
