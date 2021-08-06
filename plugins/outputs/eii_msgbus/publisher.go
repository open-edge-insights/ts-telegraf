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
	eiimsgbus "EIIMessageBus/eiimsgbus"
	"fmt"
	"time"
	"strconv"
)

// Creates new eii messagebus client handle
func (pluginPubObj *pluginPublisher) createClient() error {
	var err error
	pluginPubObj.msgBusClient, err = eiimsgbus.NewMsgbusClient(pluginPubObj.eiiMsgBusConfigMap)
	if err != nil {
		return fmt.Errorf("-- Error creating context: %v", err)
	}

	return nil
}

// Creates the eii message bus config from plugin config object
func (pluginPubObj *pluginPublisher) initEiiMsgBusConfigMap() error {

	pubCtx, err := pluginPubObj.confMgr.GetPublisherByName(pluginPubObj.pluginConfigObj.instanceName)
	if err != nil {
		return fmt.Errorf("Failed to get the publisher: %v", err)
	}

	pluginPubObj.eiiMsgBusConfigMap, err = pubCtx.GetMsgbusConfig()
	if err != nil {
		return fmt.Errorf("Error while getting eii message bus config: %v\n", err)
	}
	pluginPubObj.Log.Debugf("Plugin config is %v", pluginPubObj.eiiMsgBusConfigMap)

	return nil
}

// start publisher
func (pluginPubObj *pluginPublisher)StartPublisher(topic string) error {
	var err error
	pluginPubObj.msgBusPubMap[topic], err = pluginPubObj.msgBusClient.NewPublisher(topic)
	if err != nil {
		fmt.Errorf("-- Error creating publisher: %v\n", err)
	}
	return nil
}

// StopAllPublisher function will stop all the registered publishers
func (pluginPubObj *pluginPublisher) StopAllPublisher() {
	for _, pub := range pluginPubObj.msgBusPubMap {
		pub.Close()
	}
}

// StopClient function will stop the eii messagebus client
func (pluginPubObj *pluginPublisher) StopClient() {
	pluginPubObj.msgBusClient.Close()
}

// Publish data to message bus
func (pluginPubObj *pluginPublisher) write(topic string, data map[string]interface{}) error {
	var pub *eiimsgbus.Publisher
	var ok bool
	if pluginPubObj.pluginConfigObj.profiling {
		tsTemp := strconv.FormatInt((time.Now().UnixNano()/1e6), 10)
		data["ts_telegraf_output_pub_exit"] = tsTemp
	}
	if pluginPubObj.pluginConfigObj.measurements[0] != "*" {
		pub, ok = pluginPubObj.msgBusPubMap[topic]
	} else {
		pub, ok = pluginPubObj.msgBusPubMap[topic]
		if !ok {
			pluginPubObj.StartPublisher(topic)
			pub, ok = pluginPubObj.msgBusPubMap[topic]
		}
	}

	if ok {
			fmt.Printf("Published message: %v\n", data)
			pub.Publish(data)
		}
	return nil
}
