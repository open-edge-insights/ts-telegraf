# Dockerfile for Telegraf

ARG IEI_VERSION
FROM ia_pybase:$IEI_VERSION
LABEL description="Telegraf image"
RUN mkdir -p /IEI/telegraf_logs

# Getting Telegraf binary
RUN wget https://dl.influxdata.com/telegraf/releases/telegraf_1.9.0-1_amd64.deb && \
    dpkg -i telegraf_1.9.0-1_amd64.deb && \
    rm telegraf_1.9.0-1_amd64.deb

ENV PYTHONPATH ${PYTHONPATH}:.

ARG IEI_UID
RUN mkdir -p /etc/ssl/ca && \
    chown -R ${IEI_UID} /etc/ssl/ && \
    chown -R ${IEI_UID} /IEI/telegraf_logs 
    

ADD Telegraf/telegraf_requirements.txt . 
RUN pip3.6 install -r telegraf_requirements.txt && \
    rm -rf telegraf_requirements.txt

# Add custom python entrypoint script to get cofig and set envirnoment variable
COPY Util/ ./Util/
COPY libs/ConfigManager ./libs/ConfigManager
COPY libs/common ./libs/common
COPY Telegraf ./Telegraf

ENTRYPOINT ["python3.6","Telegraf/telegraf_start.py", "--log-dir", "/IEI/telegraf_logs"]
CMD ["--log", "DEBUG"]
HEALTHCHECK NONE

