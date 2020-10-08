/*
Copyright (c) 2020 Intel Corporation.

Permission is hereby granted, free of charge, to any person obtaining a copy of
this software and associated documentation files (the "Software"), to deal in
the Software without restriction, including without limitation the rights to
use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies
of the Software, and to permit persons to whom the Software is furnished to do
so, subject to the following conditions:

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

package eis_msgbus

import (
	eiscfgmgr "ConfigMgr/eisconfigmgr"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
	"github.com/influxdata/telegraf/plugins/parsers"
	"sync"
)

// EisMsgbus : plugin object
type EisMsgbus struct {
	Instance_name   string `toml:"instance_name"`
	pluginConfigObj eisMsgbusInputPluginConfig
	pluginRtData    pluginRuntimeData
	parser          parsers.Parser
	pluginSubObj    pluginSubscriber
	ac              telegraf.Accumulator
	Log             telegraf.Logger
	confMgr         *eiscfgmgr.ConfigMgr
}

func (emb *EisMsgbus) initPluginRtData() {
	emb.pluginRtData.tpRtData = make(map[string]*tpRuntimeData)
	emb.pluginRtData.dataChannelsOfAllTps = make(map[string]chan dataFromMsgBus)
	emb.pluginRtData.mapOfThreadPools = make(map[string]*threadPool)
	emb.pluginRtData.subControlChannel = make(chan byte)
	emb.pluginRtData.ac = emb.ac
	emb.pluginRtData.parser = emb.parser
	emb.pluginRtData.mutex = &sync.Mutex{}
}

// Description : A short description for a plugin
func (emb *EisMsgbus) Description() string {
	return "Subscriber for EIS topics"
}

const sampleConfig = `
# Most of the configuration for this plugin lies in ETCD
#
#
# Below is the sample configuration which need to be in ETCD
# {
#   "Telegraf/config":{
#      "publisher1":{
#         "topics_info":[
#            "topic-pfx1:temperature:10:2",
#            "topic-pfx2:pressure::",
#            "topic-pfx3:humidity"
#         ],
#         "queue_len":10,
#         "num_worker":2,
#         "profiling":"true"
#      }
#   },
#   "Telegraf/interfaces":{
#      "Subscribers":[
#         {
#            "Name":"publisher1",
#            "EndPoint":"127.0.0.1:5569",
#            "Topics":[
#               "topic-pfx1",
#               "topic-pfx2",
#               "topic-pfx3"
#            ],
#            "Type":"zmq_tcp"
#         }
#      ]
#   }
#}
# 
# 
# Description of the configuration is:
# Telegraf/interfaces is a messagebus configuration
#
# Telegraf/config is an App configuration
# For plugin instance named "publishe1" can be found at
#
#    "Telegraf/config":{
#      "publisher1": {
#      .
#      .
#      }
#
##The format to mention the topic are
# Option 1. ${eis-msg-topic-prefix}:${measurement-name}:${queue_len}:${num_of_workers_in_pool}
# Dedicate queue and pool of workers for a topic
# Asynchronously processing the messages
#
#
# Option 2. ${eis-msg-topic-name}:${measurement-name}::
# No deicated queue+worker-pool for a topic.
# All the messages will go to global queue and global pool of workers will process it
# Asynchronously processing the messages
#
#
# Option 3. ${eis-msg-topic-name}:${measurement-name}
# No queue + no workers =>  receive + json-processing
# MQTT input plugin does this way
# Synchronously processing the messages

# Prefix of topics to subscribe. The actual topic name can be part of topic_prefix
#topics_info = [
#   "topic-pfx1:temperature:10:2",
#   "topic-pfx2:pressure::",
#   "topic-pfx3:humidity"
#]
#
#global queue length
#queue_len = 10
#
#num of workers working on records from global queue
#num_worker = 1
#
# default value False
# Will add in/out time in the different components of plugins
#profiling = false
#
#
#Below is a config section for a plugin in telegraf.conf file
instance_name = "publisher1"
#
## Data format to consume.
## Each data format has its own unique set of configuration options, read
## more about them here:
## https://github.com/influxdata/telegraf/tree/master/plugins/parsers/json

data_format = "json"
json_strict = true
#json_query = ""
tag_keys = [
  "my_tag_1",
  "my_tag_2"
]

json_string_fields = [
  "field1",
  "field2"
]

json_name_key = ""

#json_time_key = "" 
#json_time_format = ""
#json_timezone = ""
`

// SampleConfig : Will be called by telegraf engine
func (emb *EisMsgbus) SampleConfig() string {
	return sampleConfig
}

// Gather : Will be called by telegraf engine
func (emb *EisMsgbus) Gather(acc telegraf.Accumulator) error {
	return nil
}

// Start : Will be called by telegraf engine and starts the plugin as service
func (emb *EisMsgbus) Start(ac telegraf.Accumulator) error {
	confMgr, err := eiscfgmgr.ConfigManager()
	if err != nil {
		emb.Log.Errorf(err.Error())
		return err
	}
	emb.confMgr = confMgr

	emb.ac = ac
	emb.pluginSubObj.pluginConfigObj = &(emb.pluginConfigObj)
	emb.pluginSubObj.pluginRtData = &(emb.pluginRtData)
	emb.pluginSubObj.Log = emb.Log
	emb.pluginSubObj.confMgr = emb.confMgr

	emb.initPluginRtData()

	if err := emb.readConfig(); err != nil {
		emb.Log.Errorf(err.Error())
		return err
	}
	if err := emb.initEisBus(); err != nil {
		emb.Log.Errorf(err.Error())
		return err
	}
	return nil
}

// SetParser : Will be called by telegraf engine and it sets the parser
func (emb *EisMsgbus) SetParser(parser parsers.Parser) {
	emb.parser = parser
}

// Stop : Will be called by telegraf engine and it stops all threads in plugin
func (emb *EisMsgbus) Stop() {

	emb.Log.Infof("Shutting down subscriber")
	emb.pluginSubObj.sendShutdownSignal()
	emb.pluginSubObj.waitForShutdown()
	emb.Log.Infof("Subscriber shutdown successfully")

	for name, pool := range emb.pluginRtData.mapOfThreadPools {
		emb.Log.Infof("shutting down threadpool:%v", name)
		pool.sendShutdownSignal()
	}

	for name, pool := range emb.pluginRtData.mapOfThreadPools {
		pool.waitForShutdown()
		emb.Log.Infof("threadpool:%v shutdown successfully", name)
	}
}

// 1.Creates the messagebus config and messagebus client and
//   subscrtiber handles for each configured prefix.
// 2.All subscribers goes into receive loop
func (emb *EisMsgbus) initEisBus() error {
	// Create messagebus config
	if err := emb.pluginSubObj.initEisMsgBusConfigMap(); err != nil {
		return err
	}

	// Create messagebus client
	if err := emb.pluginSubObj.createClient(); err != nil {
		return err
	}

	// Create messagebus subscriber handle
	if err := emb.pluginSubObj.subcribeToAllTopics(); err != nil {
		return err
	}

	// Subscriber goes into receive mode
	if err := emb.pluginSubObj.receiveFromAllTopics(); err != nil {
		return err
	}

	return nil
}

// Convert the pkugin configuration into eisMsgbusInputPluginConfig object
func (emb *EisMsgbus) readConfig() error {
	if err := emb.pluginConfigObj.initConfig(emb); err != nil {
		return err
	}
	emb.Log.Infof("Configuration parsed successfully :%v", emb.pluginConfigObj)
	return nil
}

func init() {
	inputs.Add("eis_msgbus", func() telegraf.Input {
		return &EisMsgbus{}
	})
}
