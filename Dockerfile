# Dockerfile for Telegraf

ARG IEI_VERSION
FROM ia_gopybase:$IEI_VERSION
LABEL description="Telegraf image"
RUN mkdir -p ${GO_WORK_DIR}/log
RUN pip install grpcio
RUN pip install grpcio-tools

# Getting Telegraf binary
RUN wget https://dl.influxdata.com/telegraf/releases/telegraf_1.9.0-1_amd64.deb && \
    dpkg -i telegraf_1.9.0-1_amd64.deb && \
    rm telegraf_1.9.0-1_amd64.deb

# Installing cryptography module
RUN pip3.6 install cryptography==2.4.2

# Add custom python entrypoint script to get cofig and set envirnoment variable
ADD Telegraf ./Telegraf
ADD DataAgent/da_grpc/ ./DataAgent/da_grpc
ADD Util/ ./Util

RUN mkdir -p /etc/ssl/ca
ENV PYTHONPATH ${PYTHONPATH}:./DataAgent/da_grpc/protobuff/py:./DataAgent/da_grpc/protobuff/py/pb_internal

ARG IEI_UID
RUN chown -R ${IEI_UID} /etc/ssl/ 
ENTRYPOINT ["python3.6","Telegraf/telegraf_start.py", "--log-dir", "/IEI/telegraf_logs"]
CMD ["--log", "DEBUG"]
HEALTHCHECK NONE
