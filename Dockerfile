FROM golang:1.12 as builder

ARG CODECOV_TOKEN
WORKDIR /app
COPY . .

RUN make iot \
 && curl -s https://codecov.io/bash | bash \
 && curl -s -o /app/cacert.pem https://curl.haxx.se/ca/cacert.pem

FROM scratch

EXPOSE 1080

HEALTHCHECK --retries=10 CMD [ "/iot", "-url", "http://localhost:1080/health" ]
ENTRYPOINT [ "/iot" ]

ARG VERSION
ENV VERSION=${VERSION}

COPY templates/ /templates
COPY static/ /static
COPY --from=builder /app/cacert.pem /etc/ssl/certs/ca-certificates.crt
COPY --from=builder /app/bin/iot /
