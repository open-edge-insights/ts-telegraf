/*
Copyright (c) 2021 Intel Corporation

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
*/

package eii_msgbus

import (
    eiicfgmgr "github.com/open-edge-insights/eii-configmgr-go/eiiconfigmgr"
    eiimsgbus "github.com/open-edge-insights/eii-messagebus-go/eiimsgbus"
    "github.com/influxdata/telegraf"
    "github.com/influxdata/telegraf/plugins/outputs"
    "github.com/influxdata/telegraf/plugins/serializers"
    "time"
    "strconv"
)


type EiiMsgbus struct {
    InstanceName   string `toml:"instance_name"`
    pluginConfigObj eiiMsgbusOutputPluginConfig
    pluginPubObj    pluginPublisher
    Log             telegraf.Logger
    confMgr         *eiicfgmgr.ConfigMgr
    serializer      serializers.Serializer
}


// Description : A short description for a plugin
func (emb *EiiMsgbus) Description() string {
    return "Publisher for EII topics"
}

const sampleConfig = `
# Most of the configuration for this plugin lies in ETCD
#
#
# Below is the sample configuration which need to be in ETCD
#{
#    "config": {
#        "publisher1": {
#            "profiling": "false"
#        }
#    },
#    "interfaces": {
#        "Publishers": [
#            {
#                "Name": "publisher1",
#                "Type": "zmq_tcp",
#                "EndPoint": "127.0.0.1:65077",
#                "Topics": [
#                    "point_data"
#                ],
#                "AllowedClients": [
#                    "*"
#                ]
#            }
#         ]
#   }
#}

# interfaces is a messagebus configuration
#
# config is an App configuration
# For plugin instance named "publishe1" can be found at
#
#    "Telegraf/config":{
#      "publisher1": {
#      .
#      .
#      }

##Will add in/out time in the different components of plugins
#profiling = false
#default value False
#
#
#Below is a config section for a plugin in telegraf.conf file
instance_name = "publisher1"
`

// SampleConfig : Will be called by telegraf engine
func (emb *EiiMsgbus) SampleConfig() string {
    return sampleConfig
}

// Init is for setup, and validating config.
func (emb *EiiMsgbus) Init() error {
    confMgr, err := eiicfgmgr.ConfigManager()
    if err != nil {
        emb.Log.Errorf(err.Error())
        return err
    }
    emb.confMgr = confMgr
    emb.pluginPubObj.pluginConfigObj = &(emb.pluginConfigObj)
    emb.pluginPubObj.Log = emb.Log
    emb.pluginPubObj.confMgr = emb.confMgr


    if err := emb.readConfig(); err != nil {
        emb.Log.Errorf(err.Error())
        return err
    }
    if err := emb.initEiiBus(); err != nil {
        emb.Log.Errorf(err.Error())
        return err
    }
    return nil
}


// Creates the messagebus config and messagebus client 
func (emb *EiiMsgbus) initEiiBus() error {
    emb.pluginPubObj.msgBusPubMap = make(map[string]*eiimsgbus.Publisher)
    // Create messagebus config
    if err := emb.pluginPubObj.initEiiMsgBusConfigMap(); err != nil {
        return err
    }

    // Create messagebus client
    if err := emb.pluginPubObj.createClient(); err != nil {
        return err
    }

    return nil
}

// Convert the plugin configuration into eiiMsgbusOutputPluginConfig object
func (emb *EiiMsgbus) readConfig() error {
    if err := emb.pluginConfigObj.initConfig(emb); err != nil {
        return err
    }
    emb.Log.Infof("Configuration parsed successfully :%v", emb.pluginConfigObj)
    return nil
}

func (emb *EiiMsgbus) Connect() error {
    // Make any connection required here
    if emb.pluginConfigObj.measurements[0] != "*" {
        for _, measurement := range emb.pluginConfigObj.measurements {
            emb.pluginPubObj.StartPublisher(measurement)
        }
    }
    return nil
}

func (emb *EiiMsgbus) Close() error {
    // Close any connections here.
    // Write will not be called once Close is called, so there is no need to synchronize.
    emb.pluginPubObj.StopAllPublisher()
    emb.pluginPubObj.StopClient()

    return nil
}

// Write should write immediately to the output, and not buffer writes
// (Telegraf manages the buffer for you). Returning an error will fail this
// batch of writes and the entire batch will be retried automatically.
func (emb *EiiMsgbus) Write(metrics []telegraf.Metric) error {
    for _, metric := range metrics {
        data := make(map[string]interface{})
        if emb.pluginConfigObj.profiling{
            tsTemp := strconv.FormatInt((time.Now().UnixNano()/1e6), 10)
            data["ts_telegraf_output_data_entry"] = tsTemp
        }
        topic := metric.Name()
        fields := metric.Fields()
        data["Name"] = topic
	// Converting int64 fields to int field, to make it acceptable to msgbug.
        var ivalue interface{}
        var cvalue interface{}
        for key, value := range fields {
            ivalue = value
            switch ivalue.(type) {
            case int64:
                val := int(ivalue.(int64))
                cvalue = val
            default:
                cvalue = ivalue
            }
            data[key] = cvalue
        }
        emb.pluginPubObj.write(topic, data)
    }
    return nil
}

func init() {
    outputs.Add("eii_msgbus", func() telegraf.Output { return &EiiMsgbus{} })
}
