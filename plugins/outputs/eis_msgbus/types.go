/*
Copyright (c) 2021 Intel Corporation.

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
	eismsgbus "EISMessageBus/eismsgbus"
	"github.com/influxdata/telegraf"
)

type eisMsgbusOutputPluginConfig struct {
	instanceName        string //eis messagebus plugin instance name
	profiling           bool // true when profiling is on, else false
	devmode             bool // true in case of developer mode else false
	measurements        []string //list of measurement for publishing data
}


type pluginPublisher struct {
	msgBusClient       *eismsgbus.MsgbusClient          // eis messsagebus client
	msgBusPubMap       map[string]*eismsgbus.Publisher // eis messagebus publisher handels for topic
	pluginConfigObj    *eisMsgbusOutputPluginConfig      // ref to plugin config object
	eisMsgBusConfigMap map[string]interface{}           // eis messagbus config
	Log                telegraf.Logger                  // telegraf logger object
	confMgr            *eiscfgmgr.ConfigMgr             // Config manager reference
}

type pubData struct {
	buf      []byte // blob data   
	profInfo map[string]interface{} // profiling data
}
