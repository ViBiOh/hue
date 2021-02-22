FROM vibioh/scratch

ENV HUE_PORT 1080
EXPOSE 1080

ENV ZONEINFO /zoneinfo.zip
COPY zoneinfo.zip /zoneinfo.zip
COPY ca-certificates.crt /etc/ssl/certs/ca-certificates.crt

COPY static/ /static

HEALTHCHECK --retries=10 CMD [ "/hue", "-url", "http://localhost:1080/health" ]
ENTRYPOINT [ "/hue" ]

ARG VERSION
ENV VERSION=${VERSION}

ARG TARGETOS
ARG TARGETARCH

COPY release/hue_${TARGETOS}_${TARGETARCH} /hue
