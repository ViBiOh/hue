FROM golang:1.13 as builder

WORKDIR /app
COPY . .

RUN make \
 && git diff -- *.go \
 && git diff --quiet -- *.go

ARG CODECOV_TOKEN
RUN curl -q -sS https://codecov.io/bash | bash

FROM alpine as fetcher

WORKDIR /app

RUN apk --update add curl \
 && curl -q -sS -o /app/cacert.pem https://curl.haxx.se/ca/cacert.pem \
 && curl -q -sS -o /app/zoneinfo.zip https://raw.githubusercontent.com/golang/go/master/lib/time/zoneinfo.zip

FROM scratch

ENV ZONEINFO zoneinfo.zip
EXPOSE 1080

HEALTHCHECK --retries=10 CMD [ "/iot", "-url", "http://localhost:1080/health" ]
ENTRYPOINT [ "/iot" ]

ARG APP_VERSION
ENV VERSION=${APP_VERSION}

COPY templates/ /templates
COPY static/ /static
COPY --from=fetcher /app/cacert.pem /etc/ssl/certs/ca-certificates.crt
COPY --from=fetcher /app/zoneinfo.zip /
COPY --from=builder /app/bin/iot /
