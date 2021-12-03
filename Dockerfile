# Copyright (c) 2020 Intel Corporation.

# Permission is hereby granted, free of charge, to any person obtaining a copy
# of this software and associated documentation files (the "Software"), to deal
# in the Software without restriction, including without limitation the rights
# to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
# copies of the Software, and to permit persons to whom the Software is
# furnished to do so, subject to the following conditions:

# The above copyright notice and this permission notice shall be included in
# all copies or substantial portions of the Software.

# THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
# IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
# FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
# AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
# LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
# OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
# SOFTWARE.

# Dockerfile for Telegraf

ARG EII_VERSION
ARG UBUNTU_IMAGE_VERSION

ARG ARTIFACTS="/artifacts"
FROM ia_common:$EII_VERSION as common
FROM ia_eiibase:${EII_VERSION} as base

FROM base as builder
LABEL description="Telegraf image"

WORKDIR /app

COPY . ./Telegraf

RUN mkdir /etc/Telegraf && \
    cp -r ./Telegraf/config/* /etc/Telegraf/ && \
    rm -rf ./Telegraf/config

ARG ARTIFACTS
RUN mkdir $ARTIFACTS \
          $ARTIFACTS/telegraf \
          $ARTIFACTS/bin

RUN mv Telegraf/schema.json $ARTIFACTS/telegraf && \
    mv Telegraf/telegraf_start.py $ARTIFACTS/telegraf

ARG CMAKE_INSTALL_PREFIX
ENV CMAKE_INSTALL_PREFIX=${CMAKE_INSTALL_PREFIX}
COPY --from=common ${CMAKE_INSTALL_PREFIX}/include ${CMAKE_INSTALL_PREFIX}/include
COPY --from=common ${CMAKE_INSTALL_PREFIX}/lib ${CMAKE_INSTALL_PREFIX}/lib

COPY --from=common /eii/common/libs/EIIMessageBus/go/EIIMessageBus /src/EIIMessageBus
COPY --from=common /eii/common/libs/ConfigMgr/go/ConfigMgr /src/ConfigMgr

ENV PATH="$PATH:/usr/local/go/bin" \
    PKG_CONFIG_PATH="$PKG_CONFIG_PATH:${CMAKE_INSTALL_PREFIX}/lib/pkgconfig" \
    LD_LIBRARY_PATH="${LD_LIBRARY_PATH}:${CMAKE_INSTALL_PREFIX}/lib"

# These flags are needed for enabling security while compiling and linking with cpuidcheck in golang
ENV CGO_CFLAGS="$CGO_FLAGS -I ${CMAKE_INSTALL_PREFIX}/include -O2 -D_FORTIFY_SOURCE=2 -Werror=format-security -fstack-protector-strong -fPIC" \
    CGO_LDFLAGS="$CGO_LDFLAGS -L${CMAKE_INSTALL_PREFIX}/lib -z noexecstack -z relro -z now"

ARG TELEGRAF_SOURCE_TAG
RUN cd /src/ && \
    git clone https://github.com/influxdata/telegraf.git && \
    cd telegraf && \
    git fetch --tags && \
    git checkout tags/${TELEGRAF_SOURCE_TAG} -b ${TELEGRAF_SOURCE_TAG}-branch

ENV TELEGRAF_SRC_DIR /src/telegraf

RUN cd $TELEGRAF_SRC_DIR/ && \
    go get google.golang.org/grpc@v1.26.0

COPY ./plugins/inputs/all/all.patch /tmp/all.patch
COPY ./plugins/inputs/eii_msgbus /src/telegraf/plugins/inputs/eii_msgbus

# Applying the patch to only single file.
RUN patch -p0 $TELEGRAF_SRC_DIR/plugins/inputs/all/all.go -i /tmp/all.patch && \
    rm -f /tmp/all.patch

COPY ./plugins/outputs/all/all.patch /tmp/all.patch
COPY ./plugins/outputs/eii_msgbus /src/telegraf/plugins/outputs/eii_msgbus

RUN patch -p0 $TELEGRAF_SRC_DIR/plugins/outputs/all/all.go -i /tmp/all.patch && \
    rm -f /tmp/all.patch

RUN echo "replace github.com/open-edge-insights/eii-configmgr-go => ../ConfigMgr/" >> $TELEGRAF_SRC_DIR/go.mod

RUN echo "replace github.com/open-edge-insights/eii-messagebus-go => ../EIIMessageBus/" >> $TELEGRAF_SRC_DIR/go.mod

RUN cd $TELEGRAF_SRC_DIR && \
    make && \
    cp /src/telegraf/telegraf $ARTIFACTS/bin && \
    rm -rf /src

FROM ubuntu:$UBUNTU_IMAGE_VERSION as runtime
WORKDIR /app
ARG ARTIFACTS
RUN apt update && apt install --no-install-recommends -y libcjson1 libzmq5 zlib1g && \
    rm -rf /var/lib/apt/lists/*
COPY --from=builder /etc/Telegraf/ /etc/Telegraf/
COPY --from=builder $ARTIFACTS/telegraf .
COPY --from=builder $ARTIFACTS/bin/telegraf .local/bin/

# Setting python env
RUN apt-get update && \
    apt-get install -y --no-install-recommends python3-distutils python3-minimal

ARG EII_UID
ARG EII_USER_NAME
RUN groupadd $EII_USER_NAME -g $EII_UID && \
    useradd -r -u $EII_UID -g $EII_USER_NAME $EII_USER_NAME


ARG CMAKE_INSTALL_PREFIX
ENV CMAKE_INSTALL_PREFIX=${CMAKE_INSTALL_PREFIX}
ENV PYTHONPATH $PYTHONPATH:/app/.local/lib/python3.8/site-packages:/app
COPY --from=common ${CMAKE_INSTALL_PREFIX}/lib ${CMAKE_INSTALL_PREFIX}/lib
COPY --from=common /eii/common/util/*.py util/
COPY --from=common /root/.local/lib .local/lib

RUN chown -R ${EII_UID}:${EII_UID} /tmp/ && \
    chmod -R 760 /tmp/
RUN chown -R ${EII_UID} .local/lib/python3.8
USER $EII_USER_NAME
ENV LD_LIBRARY_PATH $LD_LIBRARY_PATH:${CMAKE_INSTALL_PREFIX}/lib
ENV PATH $PATH:/app/.local/bin

HEALTHCHECK NONE

ENTRYPOINT ["python3","telegraf_start.py"]
