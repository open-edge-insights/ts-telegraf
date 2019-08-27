# Dockerfile for Telegraf

ARG EIS_VERSION
FROM ia_pybase:$EIS_VERSION as pybase
LABEL description="Telegraf image"

# Getting Telegraf binary
RUN wget https://dl.influxdata.com/telegraf/releases/telegraf_1.9.0-1_amd64.deb && \
    dpkg -i telegraf_1.9.0-1_amd64.deb && \
    rm telegraf_1.9.0-1_amd64.deb

ENV PYTHONPATH ${PYTHONPATH}:.

ARG EIS_UID
RUN mkdir -p /etc/ssl/ca && \
    chown -R ${EIS_UID} /etc/ssl/

COPY telegraf_requirements.txt . 
RUN pip3.6 install -r telegraf_requirements.txt && \
    rm -rf telegraf_requirements.txt

FROM ia_common:$EIS_VERSION as common

FROM pybase

COPY --from=common /libs ${PY_WORK_DIR}/libs
COPY --from=common /Util ${PY_WORK_DIR}/Util

# Add custom python entrypoint script to get cofig and set envirnoment variable

COPY telegraf_start.py ./Telegraf/telegraf_start.py

ENV INFLUX_SERVER localhost
ENV INFLUXDB_PORT 8086
ENV HOST_IP localhost
ENTRYPOINT ["python3.6","Telegraf/telegraf_start.py"]

HEALTHCHECK NONE

