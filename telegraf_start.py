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
from DataAgent.da_grpc.client.py.client_internal.client \
    import GrpcInternalClient
from Util.log import configure_logging, LOG_LEVELS
from distutils.util import strtobool
from Util.util import create_decrypted_pem_files

GRPC_CERTS_PATH = "/etc/ssl/grpc_int_ssl_secrets"
CLIENT_CERT = GRPC_CERTS_PATH + "/grpc_internal_client_certificate.pem"
CLIENT_KEY = GRPC_CERTS_PATH + "/grpc_internal_client_key.pem"
CA_CERT = GRPC_CERTS_PATH + "/ca_certificate.pem"


def parse_args():
    """Parse command line arguments
    """
    parser = argparse.ArgumentParser()
    parser.add_argument('--log', choices=LOG_LEVELS.keys(), default='INFO',
                        help='Logging level (df: INFO)')
    parser.add_argument('--log-dir', dest='log_dir', default='logs',
                        help='Directory to for log files')
    # parser.add_argument('--dev-mode', dest='dev_mode', default='False',
    #                     help='developement mode True/False')

    return parser.parse_args()


if __name__ == '__main__':

    # Parse command line arguments
    args = parse_args()
    devMode = bool(strtobool(os.environ['DEV_MODE']))

    currentDateTime = str(datetime.datetime.now())
    listDateTime = currentDateTime.split(" ")
    currentDateTime = "_".join(listDateTime)
    logFileName = 'telegraf' + currentDateTime + '.log'

    log = configure_logging(args.log.upper(), logFileName, args.log_dir,
                            __name__)

    log.info("=============== STARTING telegraf ===============")
    try:
        if not devMode:
            client = GrpcInternalClient(CLIENT_CERT, CLIENT_KEY, CA_CERT)
            srcFiles = [CA_CERT]
            filesToDecrypt = ["/etc/ssl/ca/ca_certificate.pem"]
            create_decrypted_pem_files(srcFiles, filesToDecrypt)
        else:
            client = GrpcInternalClient()

        config = client.GetConfigInt("InfluxDBCfg")
        os.environ["INFLUXDB_USERNAME"] = config["UserName"]
        os.environ["INFLUXDB_PASSWORD"] = config["Password"]
        os.environ["INFLUXDB_DBNAME"] = config["DBName"]

        if devMode:
            Telegraf_conf = "/etc/Telegraf/telegraf_devmode.conf"
        else:
            Telegraf_conf = "/etc/Telegraf/telegraf.conf"

        subprocess.call(["telegraf", "-config=" + Telegraf_conf])
    except Exception as e:
        log.error(e, exc_info=True)
        os._exit(1)
