FROM scratch

EXPOSE 1080

COPY templates/ /templates

HEALTHCHECK --retries=10 CMD [ "/hue", "-url", "http://localhost:1080/health" ]
ENTRYPOINT [ "/hue" ]

ARG VERSION
ENV VERSION=${VERSION}

ARG TARGETOS
ARG TARGETARCH

COPY cacert.pem /etc/ssl/certs/ca-certificates.crt
COPY zoneinfo.zip /
COPY release/hue_${TARGETOS}_${TARGETARCH} /hue
