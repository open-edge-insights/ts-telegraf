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
	json "encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
)

// Retuns the config object for a given topic
func (pluginConfigObj *eisMsgbusInputPluginConfig) getPrefixConfigForTopic(tpName string) *topicPrefixConfig {

	pxfLen := 0
	var config *topicPrefixConfig = nil

	// Match the longest prefix for the topic name
	for tpPrefix, tempConfigObj := range pluginConfigObj.mapOfPrefixToConfig {
		if strings.HasPrefix(tpName, tpPrefix) {
			if len(tpPrefix) > pxfLen {
				pxfLen = len(tpPrefix)
				config = tempConfigObj
			}
		}
	}

	return config
}

// Converts the plugin configuration into plugin configuration object of type pluginConfigObj.
func (pluginConfigObj *eisMsgbusInputPluginConfig) initConfig(emb *EisMsgbus) error {

	appConfig, err := emb.confMgr.GetAppConfig()
	if err != nil {
		return fmt.Errorf("Error in getting config: %v", err)
	}

	var instanceConfig map[string]interface{}
	var found bool
	if instanceConfig, found = appConfig[emb.Instance_name].(map[string]interface{}); found == false {
		return fmt.Errorf("Could not get the configuration for %v: %v", emb.Instance_name, err)
	}

	pluginConfigObj.mapOfPrefixToConfig = make(map[string]*topicPrefixConfig)
	for _, tpPfxConfLine := range instanceConfig["topics_info"].([]interface{}) {
		// Parse each prefix config line and create an prefix config object
		tempArray := strings.Split(tpPfxConfLine.(string), ":")
		tempObj := getTopicPrefixConfig(tempArray)
		if tempObj == nil {
			return fmt.Errorf("wrong prefix configuration:%v", tpPfxConfLine)
		}
		pluginConfigObj.mapOfPrefixToConfig[tempObj.tpPrefix] = tempObj
	}

	numInt, err := instanceConfig["queue_len"].(json.Number).Int64()
	if err != nil {
		return fmt.Errorf("json number conversion failed %v: %v", instanceConfig["queue_len"], err)
	}

	pluginConfigObj.globalQueueLen = int(numInt)

	numInt, err = instanceConfig["num_worker"].(json.Number).Int64()
	if err != nil {
		return fmt.Errorf("json number conversion failed %v: %v", instanceConfig["num_worker"], err)
	}
	pluginConfigObj.globalPoolSize = int(numInt)

	value := instanceConfig["profiling"].(string)
	pluginConfigObj.profiling, err = strconv.ParseBool(value)
	if err != nil {
		return fmt.Errorf("Parsing profiling mode failed: %v", err)
	}
	value = os.Getenv("DEV_MODE")
	pluginConfigObj.devmode, err = strconv.ParseBool(value)
	if err != nil {
		return fmt.Errorf("Parsing dev mode failed: %v", err)
	}

	pluginConfigObj.instanceName = emb.Instance_name

	return nil
}

// Converts the topic-prefix specific configuration into an object of type topicPrefixConfig.
func getTopicPrefixConfig(tempArray []string) *topicPrefixConfig {

	// The topic prefix configuration can be in three different forms
	// Option 1. ${eis-msg-topic-prefix}:${measurement-name}:${queue_len}:${num_of_workers_in_pool}
	// Option 2. ${eis-msg-topic-name}:${measurement-name}::
	// Option 3. ${eis-msg-topic-name}:${measurement-name}
	// All three parsing scenarios has been addressed in this function.
	if len(tempArray) < 2 || len(tempArray) > 4 {
		return nil
	}

	tpPrefix := strings.TrimSpace(tempArray[0])
	mName := strings.TrimSpace(tempArray[1])

	if len(tpPrefix) <= 0 || len(mName) <= 0 {
		return nil
	}

	obj := new(topicPrefixConfig)
	obj.isSyncProc = false
	obj.tpPrefix = tpPrefix
	obj.mName = mName

	if len(tempArray) == 2 {
		// Option 3.
		(*obj).isSyncProc = true
		return obj
	}

	queueLenInStr := strings.TrimSpace(tempArray[2])
	poolSizeInStr := strings.TrimSpace(tempArray[3])

	if len(queueLenInStr) > 0 && len(poolSizeInStr) > 0 {
		// Option 1.
		if queueLen, err := strconv.Atoi(queueLenInStr); err == nil {
			obj.queueLen = queueLen
		} else {
			return nil
		}

		if poolSize, err := strconv.Atoi(poolSizeInStr); err == nil {
			obj.poolSize = poolSize
		} else {
			return nil
		}
	}
	// Option 2. in case of above if block skipped

	return obj
}
