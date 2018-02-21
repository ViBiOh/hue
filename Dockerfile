FROM scratch

HEALTHCHECK --retries=10 CMD [ "/iot", "-url", "https://localhost:1080/health" ]

ENTRYPOINT [ "/iot" ]
EXPOSE 1080

COPY cacert.pem /etc/ssl/certs/ca-certificates.crt
COPY web/ /web
COPY bin/iot /iot
