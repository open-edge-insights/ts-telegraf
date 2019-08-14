# Dockerfile for Telegraf

ARG EIS_VERSION
FROM ia_pybase:$EIS_VERSION
LABEL description="Telegraf image"
RUN mkdir -p /EIS/telegraf_logs

# Getting Telegraf binary
RUN wget https://dl.influxdata.com/telegraf/releases/telegraf_1.9.0-1_amd64.deb && \
    dpkg -i telegraf_1.9.0-1_amd64.deb && \
    rm telegraf_1.9.0-1_amd64.deb

ENV PYTHONPATH ${PYTHONPATH}:.

ARG EIS_UID
RUN mkdir -p /etc/ssl/ca && \
    chown -R ${EIS_UID} /etc/ssl/ && \
    chown -R ${EIS_UID} /EIS/telegraf_logs 
    

ADD telegraf_requirements.txt . 
RUN pip3.6 install -r telegraf_requirements.txt && \
    rm -rf telegraf_requirements.txt

# Add custom python entrypoint script to get cofig and set envirnoment variable

ADD telegraf_start.py ./Telegraf/telegraf_start.py

ENV INFLUX_SERVER localhost
ENV INFLUXDB_PORT 8086
ENV HOST_IP localhost
ENTRYPOINT ["python3.6","Telegraf/telegraf_start.py", "--log-dir", "/EIS/telegraf_logs"]

HEALTHCHECK NONE

