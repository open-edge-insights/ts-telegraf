#!/usr/bin/python3

"""
Copyright (c) 2018 Intel Corporation.
Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:
The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.
THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
"""

import os
import datetime
import time
import argparse
import logging
import subprocess
import json
from libs.ConfigManager import ConfigManager
from util.log import configure_logging, LOG_LEVELS
from distutils.util import strtobool

ETCD_CERT = "/run/secrets/etcd_InfluxDBConnector_cert"
ETCD_KEY = "/run/secrets/etcd_InfluxDBConnector_key"
CA_CERT = "/run/secrets/ca_etcd"
INFLUX_CA_KEY = "/InfluxDBConnector/ca_cert"
INFLUX_CA_PATH = "/etc/ssl/ca/ca_certificate.pem"


def read_config(client, dev_mode):
    """Read the configuration from etcd
    """
    key = ETCD_CERT.split('/')
    app_name = key[3].split('_')
    config_key_path = "config"
    configfile = client.GetConfig("/{0}/{1}".format(
                 app_name[1], config_key_path))
    config = json.loads(configfile)
    os.environ["INFLUXDB_USERNAME"] = config["influxdb"]["username"]
    os.environ["INFLUXDB_PASSWORD"] = config["influxdb"]["password"]
    os.environ["INFLUXDB_DBNAME"] = config["influxdb"]["dbname"]

    if not dev_mode:
        cert = client.GetConfig(INFLUX_CA_KEY)
        try:
            with open(INFLUX_CA_PATH, 'wb+') as fd:
                fd.write(cert.encode())
        except Exception as e:
            log.debug("Failed creating file: {}, Error: {} ".format(INFLUX_CA_PATH,
                                                                    e))

if __name__ == '__main__':
    dev_mode = bool(strtobool(os.environ["DEV_MODE"]))
    # Initializing Etcd to set env variables
    conf = {
        "certFile": "",
        "keyFile": "",
        "trustFile": ""
    }
    if not dev_mode:
        conf = {
            "certFile": ETCD_CERT,
            "keyFile": ETCD_KEY,
            "trustFile": CA_CERT
        }
    cfg_mgr = ConfigManager()
    config_client = cfg_mgr.get_config_client("etcd", conf)

    log = configure_logging(os.environ['PY_LOG_LEVEL'].upper(),
                            __name__,dev_mode)


    log.info("=============== STARTING telegraf ===============")
    try:
        if dev_mode:
            Telegraf_conf = "/etc/Telegraf/telegraf_devmode.conf"
        else:
            Telegraf_conf = "/etc/Telegraf/telegraf.conf"

        read_config(config_client, dev_mode)
        subprocess.call(["telegraf", "-config=" + Telegraf_conf])
    except Exception as e:
        log.error(e, exc_info=True)
        os._exit(1)
