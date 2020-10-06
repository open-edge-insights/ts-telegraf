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
	eismsgbus "EISMessageBus/eismsgbus"
	"fmt"
	"time"
)

// Creates new eis messagebus client handle
func (pluginSubObj *pluginSubscriber) createClient() error {
	var err error
	pluginSubObj.msgBusClient, err = eismsgbus.NewMsgbusClient(pluginSubObj.eisMsgBusConfigMap)
	if err != nil {
		return fmt.Errorf("-- Error creating context: %v", err)
	}

	return nil
}

// Creates the eis message bus config from plugin config object
func (pluginSubObj *pluginSubscriber) initEisMsgBusConfigMap() error {
	subCtx, err := pluginSubObj.confMgr.GetSubscriberByName(pluginSubObj.pluginConfigObj.instanceName)
	if err != nil {
		return fmt.Errorf("Failed to get the subscriber: %v", err)
	}

	pluginSubObj.eisMsgBusConfigMap, err = subCtx.GetMsgbusConfig()
	if err != nil {
		return fmt.Errorf("Error while getting eis message bus config: %v\n", err)
	}

	pluginSubObj.Log.Debugf("Plugin config is %v", pluginSubObj.eisMsgBusConfigMap)

	return nil
}

// Creates the subscriber for each topic prefix
func (pluginSubObj *pluginSubscriber) subcribeToAllTopics() error {
	var err error
	pluginSubObj.msgBusSubscriber, err = pluginSubObj.msgBusClient.NewSubscriber("")
	if err != nil {
		return fmt.Errorf("Error creating subscriber:%v", err)
	}
	return nil
}

// Will put each subscriber into receiving mode.
func (pluginSubObj *pluginSubscriber) receiveFromAllTopics() error {
	pluginSubObj.pluginRtData.subControlChannel = make(chan byte)
	pluginSubObj.wg.Add(1)
	go pluginSubObj.processMsg()
	return nil
}

// Send shutdown signal for every thread
func (pluginSubObj *pluginSubscriber) sendShutdownSignal() {
	pluginSubObj.pluginRtData.subControlChannel <- 'E'
}

// Wait for shutdown to happen
func (pluginSubObj *pluginSubscriber) waitForShutdown() {
	pluginSubObj.wg.Wait()
}

// Creates the topic specific runtime data
func (pluginSubObj *pluginSubscriber) createTpRtDataObject(config *topicPrefixConfig, tpName string) *tpRuntimeData {

	tpRtInfo := new(tpRuntimeData)
	tpRtInfo.mName = (*config).mName
	tpRtInfo.isSyncProc = (*config).isSyncProc
	tpRtInfo.tpName = tpName

	tpRtInfo.writer = telegrafAccWriter{ac: pluginSubObj.pluginRtData.ac}
	tpRtInfo.parser = pluginSubObj.pluginRtData.parser

	return tpRtInfo
}

// Creates the topic specific runtime data and processing pool (if configured)
func (pluginSubObj *pluginSubscriber) createRtDataAndPool(pfConfig *topicPrefixConfig, tpName string) *tpRuntimeData {

	tpRtInfo := pluginSubObj.createTpRtDataObject(pfConfig, tpName)
	pluginSubObj.pluginRtData.tpRtData[tpName] = tpRtInfo

	createPool := false
	globalPool := false
	queueLen := 0
	poolSize := 0
	var dataChannel chan dataFromMsgBus

	if (*pfConfig).isSyncProc == false {
		// Async processing
		if (*pfConfig).poolSize != 0 {
			// topic specific queue and pool
			queueLen = pfConfig.queueLen
			poolSize = pfConfig.poolSize
			createPool = true
		} else if pluginSubObj.pluginRtData.dataChannel == nil {
			// Create the global queue and pool only once
			queueLen = pluginSubObj.pluginConfigObj.globalQueueLen
			poolSize = pluginSubObj.pluginConfigObj.globalPoolSize
			createPool = true
			globalPool = true
		} else {
			// use exiting global queue and pool
			dataChannel = pluginSubObj.pluginRtData.dataChannel
		}

		if dataChannel == nil {
			dataChannel = make(chan dataFromMsgBus, queueLen)
		}
	}

	pluginSubObj.pluginRtData.dataChannelsOfAllTps[tpName] = dataChannel
	tpRtInfo.dataChannel = dataChannel

	if createPool == true {
		poolName := ""
		if globalPool == false {
			poolName = "for-" + tpName
			pluginSubObj.Log.Infof("Launching thread pool of size %v for a topic %v", poolSize, tpName)
		} else {
			poolName = "GLOBAL"
			pluginSubObj.Log.Infof("Launching global thread pool of size %v, topic detected is %v", poolSize, tpName)
		}

		var simpleProc simpleMsgProcessor
		var pool threadPool
		pool.initThrPool(&simpleProc, tpRtInfo, poolSize, pluginSubObj.Log)
		pool.setName(poolName)
		pluginSubObj.pluginRtData.mapOfThreadPools[poolName] = &pool
		pool.start()
	}

	return tpRtInfo
}

// Receiving loop of subscriber
func (pluginSubObj *pluginSubscriber) processMsg() {
	defer pluginSubObj.wg.Done()

	pluginSubObj.Log.Infof("Started subscriber's receive loop")
	var tpRtInfo *tpRuntimeData
	var found bool
	var dataChannel chan<- dataFromMsgBus

	loop := true
	for loop {
		select {

		case msg := <-pluginSubObj.msgBusSubscriber.MessageChannel:
			pluginSubObj.Log.Debugf("Data received for topic:%s", msg.Name)
			var profInfo map[string]interface{}

			if pluginSubObj.pluginConfigObj.profiling {
				profInfo = make(map[string]interface{})
				profInfo["ts_plugin_in"] = time.Now().UnixNano()
			}

			if tpRtInfo, found = pluginSubObj.pluginRtData.tpRtData[msg.Name]; found == false {
				// New topic detected.
				// Do prefix match and get the configuration for a topic prefix
				// Create the new topic specific runtime data.
				pfConfig := pluginSubObj.pluginConfigObj.getPrefixConfigForTopic(msg.Name)
				if pfConfig == nil {
					pluginSubObj.Log.Errorf("No prefix for a topic: %v, ignoring the data", msg.Name)
					continue
				}

				tpRtInfo = pluginSubObj.createRtDataAndPool(pfConfig, msg.Name)
			}

			if tpRtInfo.isSyncProc == true {
				// synchronous processing. Directly invoke the processor function
				pluginSubObj.Log.Infof("Processing the data, synchronously for topic:%v", msg.Name)
				var simpleProc simpleMsgProcessor
				err := simpleProc.processData(tpRtInfo, dataFromMsgBus{msg: msg, profInfo: profInfo})
				if err != nil {
					pluginSubObj.Log.Errorf(err.Error())
				}
			} else {
				// Asynchronous processing : Put data into respective queue
				pluginSubObj.Log.Infof("Processing the data, asynchronously for topic:%v", msg.Name)
				dataChannel = pluginSubObj.pluginRtData.dataChannelsOfAllTps[msg.Name]
				if profInfo != nil {
					profInfo["ts_queue_in"] = time.Now().UnixNano()
				}
				dataChannel <- dataFromMsgBus{msg: msg, profInfo: profInfo}
			}

		case err := <-pluginSubObj.msgBusSubscriber.ErrorChannel:
			fmt.Printf("Error received from channel \n")
			pluginSubObj.Log.Errorf("Error receiving message:%v", err)

		case exit := <-pluginSubObj.pluginRtData.subControlChannel:
			pluginSubObj.Log.Infof("Subscriber received exit signal %v\n", exit)
			loop = false

		}
	}

	pluginSubObj.msgBusSubscriber.Close()
	pluginSubObj.Log.Infof("Exiting subscriber's receive loop")
}
