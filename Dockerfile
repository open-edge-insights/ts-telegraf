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

ARG EIS_VERSION
ARG DOCKER_REGISTRY
FROM ${DOCKER_REGISTRY}ia_eisbase:$EIS_VERSION as eisbase
LABEL description="Telegraf image"

ENV PYTHONPATH ${PYTHONPATH}:.

ARG EIS_UID
RUN mkdir -p /etc/ssl/ca && \
    chown -R ${EIS_UID} /etc/ssl/

ARG DOCKER_REGISTRY
FROM ${DOCKER_REGISTRY}ia_common:$EIS_VERSION as common

FROM eisbase

COPY --from=common ${GO_WORK_DIR}/common/libs ${PY_WORK_DIR}/libs
COPY --from=common ${GO_WORK_DIR}/common/util ${PY_WORK_DIR}/util
COPY --from=common ${GO_WORK_DIR}/common/cmake ${PY_WORK_DIR}/common/cmake
COPY --from=common /usr/local/lib /usr/local/lib
COPY --from=common /usr/local/include /usr/local/include
COPY --from=common /usr/local/lib/python3.6/dist-packages/ /usr/local/lib/python3.6/dist-packages/
COPY --from=common ${GO_WORK_DIR}/common/cmake ${GO_WORK_DIR}/common/cmake

ARG TELEGRAF_GO_VERSION
RUN wget https://golang.org/dl/go${TELEGRAF_GO_VERSION}.linux-amd64.tar.gz && \
    rm -rf /usr/local/go && tar -C /usr/local -xzf go${TELEGRAF_GO_VERSION}.linux-amd64.tar.gz

COPY --from=common ${GO_WORK_DIR}/../EISMessageBus /usr/local/go/src/EISMessageBus
COPY --from=common ${GO_WORK_DIR}/../ConfigMgr /usr/local/go/src/ConfigMgr

ARG TELEGRAF_SOURCE_TAG
RUN mkdir /src/ && \
    cd /src/ && \
    git clone https://github.com/influxdata/telegraf.git && \
    cd telegraf && \
    git fetch --tags && \
    git checkout tags/${TELEGRAF_SOURCE_TAG} -b ${TELEGRAF_SOURCE_TAG}-branch

ENV TELEGRAF_SRC_DIR /src/telegraf

WORKDIR ${TELEGRAF_SRC_DIR}
RUN go get google.golang.org/grpc@v1.26.0
COPY ./plugins/inputs/all/all.patch /tmp/all.patch
COPY ./plugins/inputs/eis_msgbus /src/telegraf/plugins/inputs/eis_msgbus

# Applying the patch to only single file.
RUN patch -p0 ./plugins/inputs/all/all.go -i /tmp/all.patch && rm -f /tmp/all.patch

RUN make
RUN mv /src/telegraf/telegraf /usr/local/go/bin/telegraf
RUN rm -rf /src/telegraf

COPY ./TestPublisherApp /src/TestPublisherApp
RUN go build -o /src/TestPublisherApp/publisher /src/TestPublisherApp/publisher.go

WORKDIR /EIS
COPY . ./Telegraf/
RUN mkdir /etc/Telegraf && \
    cp -r ./Telegraf/config/* /etc/Telegraf/ && \
    rm -rf ./Telegraf/config

#Removing build dependencies
RUN apt-get remove -y wget && \
    apt-get remove -y git && \
    apt-get remove curl && \
    apt-get autoremove -y

HEALTHCHECK NONE

ENTRYPOINT ["python3.6","Telegraf/telegraf_start.py"]
