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
import sys
import subprocess
import json
import tempfile
from distutils.util import strtobool
from eis.config_manager import ConfigManager
from util.log import configure_logging
from util.util import Util


TMP_DIR = tempfile.gettempdir()
INFLUX_CA_PATH = os.path.join(TMP_DIR, "ca_certificate.pem")


def read_config(client, dev_mode, log):
    """Read the configuration from etcd
    """
    influx_app_name = os.environ["InfluxDbAppName"]
    config_key_path = "config"
    configfile = client.GetConfig("/{0}/{1}".format(influx_app_name,
                                                    config_key_path))
    config = json.loads(configfile)
    os.environ["INFLUXDB_USERNAME"] = config["influxdb"]["username"]
    os.environ["INFLUXDB_PASSWORD"] = config["influxdb"]["password"]
    os.environ["INFLUXDB_DBNAME"] = config["influxdb"]["dbname"]

    if not dev_mode:
        influx_ca_key = "/" + influx_app_name + "/ca_cert"
        cert = client.GetConfig(influx_ca_key)
        try:
            with open(INFLUX_CA_PATH, 'wb+') as fpd:
                fpd.write(cert.encode())
        except (OSError, IOError) as err:
            log.debug("Failed creating file: {}, Error: {} ".format(
                INFLUX_CA_PATH, err))


def main():
    """Main to start the telegraf service
    """
    dev_mode = bool(strtobool(os.environ["DEV_MODE"]))
    app_name = str(os.environ["AppName"])
    # Initializing Etcd to set env variables
    influx_app_name = os.environ["InfluxDbAppName"]
    conf = Util.get_crypto_dict(influx_app_name)
    cfg_mgr = ConfigManager()
    config_client = cfg_mgr.get_config_client("etcd", conf)

    log = configure_logging(os.environ['PY_LOG_LEVEL'].upper(),
                            __name__, dev_mode)

    log.info("=============== STARTING telegraf ===============")
    try:
        command = None
        if len(sys.argv) > 1:
            command = str(sys.argv[1])
        read_config(config_client, dev_mode, log)
        if command is None:
            if dev_mode:
                telegraf_conf = "/etc/Telegraf/" \
                                + app_name \
                                + "/" \
                                + app_name \
                                + "_devmode.conf"
            else:
                telegraf_conf = "/etc/Telegraf/"+app_name+"/"+app_name+".conf"
            subprocess.call(["telegraf", "-config=" + telegraf_conf])
        else:
            subprocess.call(command.split())

    except subprocess.CalledProcessError as err:
        log.error(err, exc_info=True)
        sys.exit(1)


if __name__ == '__main__':
    main()
