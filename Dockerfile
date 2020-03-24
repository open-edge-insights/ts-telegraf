# Dockerfile for Telegraf

ARG EIS_VERSION
FROM ia_eisbase:$EIS_VERSION as eisbase
LABEL description="Telegraf image"

# Getting Telegraf binary
ARG TELEGRAF_VERSION
RUN wget https://dl.influxdata.com/telegraf/releases/telegraf_${TELEGRAF_VERSION}_amd64.deb && \
    dpkg -i telegraf_${TELEGRAF_VERSION}_amd64.deb && \
    rm telegraf_${TELEGRAF_VERSION}_amd64.deb

ENV PYTHONPATH ${PYTHONPATH}:.

ARG EIS_UID
RUN mkdir -p /etc/ssl/ca && \
    chown -R ${EIS_UID} /etc/ssl/

FROM ia_common:$EIS_VERSION as common

FROM eisbase

COPY --from=common ${GO_WORK_DIR}/common/libs ${PY_WORK_DIR}/libs
COPY --from=common ${GO_WORK_DIR}/common/util ${PY_WORK_DIR}/util
COPY --from=common ${GO_WORK_DIR}/common/cmake ${PY_WORK_DIR}/common/cmake
COPY --from=common /usr/local/lib /usr/local/lib
COPY --from=common /usr/local/lib/python3.6/dist-packages/ /usr/local/lib/python3.6/dist-packages/

# Add custom python entrypoint script to get cofig and set envirnoment variable

COPY . ./Telegraf/
RUN mkdir /etc/Telegraf && \
    cp ./Telegraf/config/telegraf.conf /etc/Telegraf/ && \
    cp ./Telegraf/config/telegraf_devmode.conf /etc/Telegraf/ && \
    rm -rf ./Telegraf/config

#Removing build dependencies
RUN apt-get remove -y wget && \
    apt-get remove -y git && \
    apt-get remove curl && \
    apt-get autoremove -y

ENTRYPOINT ["python3.6","Telegraf/telegraf_start.py"]
