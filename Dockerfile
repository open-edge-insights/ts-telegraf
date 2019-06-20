# Dockerfile for Telegraf

ARG IEI_VERSION
FROM ia_pybase:$IEI_VERSION
LABEL description="Telegraf image"
RUN mkdir -p ${GO_WORK_DIR}/log

# Getting Telegraf binary
RUN wget https://dl.influxdata.com/telegraf/releases/telegraf_1.9.0-1_amd64.deb && \
    dpkg -i telegraf_1.9.0-1_amd64.deb && \
    rm telegraf_1.9.0-1_amd64.deb

ENV PYTHONPATH ${PYTHONPATH}:./DataAgent/da_grpc/protobuff/py:./DataAgent/da_grpc/protobuff/py/pb_internal
ARG IEI_UID
RUN mkdir -p /etc/ssl/ca && \
    chown -R ${IEI_UID} /etc/ssl/ 

ADD Telegraf/telegraf_requirements.txt . 
RUN pip3.6 install -r telegraf_requirements.txt && \
    rm -rf telegraf_requirements.txt

# Add custom python entrypoint script to get cofig and set envirnoment variable
COPY Util/ ./Util/
COPY Telegraf ./Telegraf
COPY DataAgent/da_grpc/ ./DataAgent/da_grpc

ENTRYPOINT ["python3.6","Telegraf/telegraf_start.py", "--log-dir", "/IEI/telegraf_logs"]
CMD ["--log", "DEBUG"]
HEALTHCHECK NONE

