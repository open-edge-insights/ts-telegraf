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
import cfgmgr.config_manager as cfg
from util.log import configure_logging


TMP_DIR = tempfile.gettempdir()
INFLUX_CA_PATH = os.path.join(TMP_DIR, "ca_certificate.pem")


def read_config(app_cfg, dev_mode, log):
    """Read the configuration from etcd
    """
    os.environ["INFLUXDB_USERNAME"] = app_cfg["influxdb"]["username"]
    os.environ["INFLUXDB_PASSWORD"] = app_cfg["influxdb"]["password"]
    os.environ["INFLUXDB_DBNAME"] = app_cfg["influxdb"]["dbname"]

    if not dev_mode:
        try:
            with open(INFLUX_CA_PATH, 'w') as fpd:
                fpd.write(app_cfg["ca_cert"])
        except (OSError, IOError) as err:
            log.debug("Failed creating file: {}, Error: {} ".format(
                INFLUX_CA_PATH, err))


def main():
    """Main to start the telegraf service
    """
    ctx = cfg.ConfigMgr()
    app_cfg = ctx.get_app_config()
    app_name = ctx.get_app_name()
    dev_mode = ctx.is_dev_mode()
    cfg_inst = os.getenv('ConfigInstance', app_name)
    log = configure_logging(os.environ['PY_LOG_LEVEL'].upper(),
                            __name__, dev_mode)

    log.info("=============== STARTING telegraf ===============")
    try:
        command = None
        if len(sys.argv) > 1:
            command = str(sys.argv[1])
        read_config(app_cfg, dev_mode, log)
        if command is None:
            if dev_mode:
                telegraf_conf = "/etc/Telegraf/" \
                                + cfg_inst \
                                + "/" \
                                + cfg_inst \
                                + "_devmode.conf"
            else:
                telegraf_conf = "/etc/Telegraf/"+cfg_inst+"/"+cfg_inst+".conf"
            subprocess.call(["telegraf", "-config=" + telegraf_conf])
        else:
            subprocess.call(command.split())

    except subprocess.CalledProcessError as err:
        log.error(err, exc_info=True)
        sys.exit(1)


if __name__ == '__main__':
    main()
