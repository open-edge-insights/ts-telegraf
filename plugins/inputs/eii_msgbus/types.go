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
	eiicfgmgr "ConfigMgr/eiiconfigmgr"
	eiimsgbus "EIIMessageBus/eiimsgbus"
	types "EIIMessageBus/pkg/types"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/parsers"
	"sync"
)

type topicPrefixConfig struct {
	tpPrefix   string // topic prefix
	mName      string // influxdb measurement name
	queueLen   int    // length of topic specific queue
	poolSize   int    // size of topic specific threadpool
	isSyncProc bool
}

type eiiMsgbusInputPluginConfig struct {
	instanceName        string //eii messagebus plugin instance name
	mapOfPrefixToConfig map[string]*topicPrefixConfig
	globalQueueLen      int  //Length of global queue
	globalPoolSize      int  // size of global threadpool
	profiling           bool // true when profiling is on, else false
	devmode             bool // true in case of developer mode else false
}

type dataFromMsgBus struct {
	msg      *types.MsgEnvelope     // Topic data from eii messagebus
	profInfo map[string]interface{} // profiling data for a point
}

// Created by producer(pluginSubscriber object)
// To be used by two types consumers
// Consumer type 1. threadpool object in case of asynch processing
// Consumer type 2. processor object in case of synch processing
type tpRuntimeData struct {
	dataChannel <-chan dataFromMsgBus // in case of async queue holding data
	mName       string                // measurement Name
	tpName      string                // topic Name
	isSyncProc  bool                  // is to process synchronously
	parser      parsers.Parser        // Json parser to be used
	writer      telegrafAccWriter
}

// To be used by producer(pluginSubscriber object)
type pluginRuntimeData struct {
	tpRtData             map[string]*tpRuntimeData
	dataChannelsOfAllTps map[string]chan dataFromMsgBus
	mapOfThreadPools     map[string]*threadPool // threadpool map specific to topic
	dataChannel          chan dataFromMsgBus    // global queue
	subControlChannel    chan byte              // controlling subscriber loops
	parser               parsers.Parser         // json parser to be used
	ac                   telegraf.Accumulator
	mutex                *sync.Mutex // lock to gaurd pluginRuntimeData
}

type pluginSubscriber struct {
	msgBusClient       *eiimsgbus.MsgbusClient          // eii messsagebus client
	msgBusSubMap       map[string]*eiimsgbus.Subscriber // eii messagebus subscriber handels for topic prefix
	subTopics          []string                         // sub topics from etcd interface
	pluginConfigObj    *eiiMsgbusInputPluginConfig      // ref to plugin config onject
	pluginRtData       *pluginRuntimeData               // ref to plugin runtime data
	eiiMsgBusConfigMap map[string]interface{}           // eii messagbus config
	Log                telegraf.Logger                  // telegraf logger object
	wg                 sync.WaitGroup                   // to waid for all subscribers to gracefully exit
	confMgr            *eiicfgmgr.ConfigMgr             // Config manager reference
}

// The wraper for Telegraf engine components
type telegrafAccWriter struct {
	ac telegraf.Accumulator
}

// An interface for any processor
type eiiMsgProcessor interface {
	processData(rtInfo *tpRuntimeData, data dataFromMsgBus) error
}

// this type implements interface eiiMsgProcessor
// It does two things
// 1.json parsing and convert the json into a metrics
// 2.Writes metrics to telegraf accumelator
type simpleMsgProcessor struct {
}

type threadPool struct {
	processor    eiiMsgProcessor // processor object
	tpRtInfo     *tpRuntimeData  // topic specifcc runtinme data
	poolSize     int             // number of threads in a pool
	Log          telegraf.Logger // telegraf logger object
	contrChannel chan byte       //channel for signals to stop all threads
	name         string          // threadpool name
	wg           sync.WaitGroup  // to waid for all thread to gracefully exit
}
