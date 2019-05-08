FROM golang:1.12 as builder

ENV APP_NAME iot

WORKDIR /app
COPY . .

RUN make ${APP_NAME} \
 && curl -s -o /app/cacert.pem https://curl.haxx.se/ca/cacert.pem

FROM scratch

ENV APP_NAME iot
EXPOSE 1080

HEALTHCHECK --retries=10 CMD [ "/iot", "-url", "http://localhost:1080/health" ]
ENTRYPOINT [ "/iot" ]

COPY templates/ /templates
COPY static/ /static
COPY --from=builder /app/cacert.pem /etc/ssl/certs/ca-certificates.crt
COPY --from=builder /app/bin/${APP_NAME} /
