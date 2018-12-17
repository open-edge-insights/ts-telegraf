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
import subprocess
from DataAgent.da_grpc.client.py.client_internal.client \
    import GrpcInternalClient

CLIENT_CERT = "/etc/ssl/grpc_int_ssl_secrets/grpc_internal_client_certificate.pem"
CLIENT_KEY = "/etc/ssl/grpc_int_ssl_secrets/grpc_internal_client_key.pem"
CA_CERT = "/etc/ssl/grpc_int_ssl_secrets/ca_certificate.pem"

if __name__ == '__main__':
    try:
        client = GrpcInternalClient(CLIENT_CERT, CLIENT_KEY, CA_CERT)
        config = client.GetConfigInt("InfluxDBCfg")
        os.environ["INFLUXDB_USERNAME"] = config["UserName"]
        os.environ["INFLUXDB_PASSWORD"] = config["Password"]
        os.environ["INFLUXDB_DBNAME"] = config["DBName"]

        subprocess.call(["telegraf", "-config=Telegraf/Telegraf.conf"])
    except Exception as e:
        print(e)
        os._exit(1)
