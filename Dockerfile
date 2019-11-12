# Dockerfile for Telegraf

ARG EIS_VERSION
FROM ia_eisbase:$EIS_VERSION as eisbase
LABEL description="Telegraf image"

# Getting Telegraf binary
RUN wget https://dl.influxdata.com/telegraf/releases/telegraf_1.9.0-1_amd64.deb && \
    dpkg -i telegraf_1.9.0-1_amd64.deb && \
    rm telegraf_1.9.0-1_amd64.deb

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

COPY telegraf_start.py ./Telegraf/telegraf_start.py

#Removing build dependencies
RUN apt-get remove -y wget && \
    apt-get remove -y git && \
    apt-get remove curl && \
    apt-get autoremove -y

ENV INFLUX_SERVER localhost
ENV INFLUXDB_PORT 8086
ENV HOST_IP localhost
ENTRYPOINT ["python3.6","Telegraf/telegraf_start.py"]
