FROM golang:1.19 AS build

COPY . /oidc-discovery

WORKDIR /oidc-discovery

ENV GOOS=linux
ENV CGO_ENABLED=0

RUN ls /oidc-discovery && \
    go get -d -v ./...

RUN go build -v -o oidc-discovery .

FROM registry.suse.com/bci/bci-micro:15.4

COPY --from=build /oidc-discovery/oidc-discovery /usr/local/bin/

RUN mkdir -p /home/oidc-discovery && \
    chown -R 1000:1000 /home/oidc-discovery && \
    echo "oidc-discovery:x:1000:1000:oidc-discovery:/tmp:/bin/bash" >> /etc/passwd && \
    echo "oidc-discovery:x:1000:" >> /etc/group

USER 1000

ENTRYPOINT ["oidc-discovery"]