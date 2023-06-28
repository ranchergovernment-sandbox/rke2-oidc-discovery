FROM cgr.dev/chainguard/go AS build

COPY . /oidc-discovery

WORKDIR /oidc-discovery

ENV GOOS=linux
ENV CGO_ENABLED=0

RUN go get -d -v ./...

RUN go build -v -o oidc-discovery .

FROM cgr.dev/chainguard/glibc-dynamic

COPY --from=build /oidc-discovery/oidc-discovery /usr/local/bin/

ENTRYPOINT ["oidc-discovery"]