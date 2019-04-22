# Dockerfile for Telegraf

ARG IEI_VERSION
FROM ia_pybase:$IEI_VERSION
LABEL description="Telegraf image"

ENV PYTHONPATH .:./DataAgent/da_grpc/protobuff:./DataAgent/da_grpc/protobuff/py:./DataAgent/da_grpc/protobuff/py/pb_internal
ENV PACKAGES="\
  libstdc++ \
  iputils \
  ca-certificates \
  net-snmp-tools \
  openssl-dev \
  "
ARG IEI_UID
ENV GO_WORK_DIR /IEI/go/src/IEdgeInsights

WORKDIR ${GO_WORK_DIR} 

RUN apk add --no-cache $PACKAGES \ 
    && mkdir -p ${GO_WORK_DIR}/log \ 
    && update-ca-certificates

RUN apk add --no-cache --virtual .build-deps \
    python3-dev \
    build-base \
    && pip3.6 install grpcio-tools

RUN mkdir -p /etc/ssl/ca && \
    chown -R ${IEI_UID} /etc/ssl/


# Installing Telegraf 
ARG TELEGRAF_VERSION
RUN set -ex && \
    apk add --no-cache --virtual .build-deps wget tar gnupg ca-certificates && \
    for key in \
        05CE15085FC09D18E99EFB22684A14CF2582E0C5 ; \
    do \
        gpg --keyserver ha.pool.sks-keyservers.net --keyserver-options http-proxy=$http_proxy --recv-keys "$key" || \
        gpg --keyserver pgp.mit.edu --keyserver-options http-proxy=$http_proxy --recv-keys "$key" || \
        gpg --keyserver keyserver.pgp.com http-proxy=$http_proxy --recv-keys "$key" ; \
    done && \
    wget -q https://dl.influxdata.com/telegraf/releases/telegraf-${TELEGRAF_VERSION}-static_linux_amd64.tar.gz.asc && \
    wget -q https://dl.influxdata.com/telegraf/releases/telegraf-${TELEGRAF_VERSION}-static_linux_amd64.tar.gz && \
    gpg --batch --verify telegraf-${TELEGRAF_VERSION}-static_linux_amd64.tar.gz.asc telegraf-${TELEGRAF_VERSION}-static_linux_amd64.tar.gz && \
    mkdir -p /usr/src /etc/telegraf && \
    tar -C /usr/src -xzf telegraf-${TELEGRAF_VERSION}-static_linux_amd64.tar.gz && \
    mv /usr/src/telegraf*/telegraf.conf /etc/telegraf/ && \
    chmod +x /usr/src/telegraf*/* && \
    cp -a /usr/src/telegraf*/* /usr/bin/ && \
    rm -rf *.tar.gz* /usr/src /root/.gnupg && \
    apk del .build-deps && \
    rm -rf /var/cache/apk/* 

# Add custom python entrypoint script to get cofig and set envirnoment variable
ADD Telegraf ./Telegraf
ADD DataAgent/da_grpc/ ./DataAgent/da_grpc
ADD Util/ ./Util
ENTRYPOINT ["python3.6","Telegraf/telegraf_start.py", "--log-dir", "/IEI/telegraf_logs"]
CMD ["--log", "DEBUG"]
HEALTHCHECK NONE
