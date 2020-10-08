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
	eismsgbustype "EISMessageBus/pkg/types"
	"encoding/json"
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

// This function returns the test json meesage
func getTestJsonData() map[string]interface{} {
	jsonMsg := map[string]interface{}{
		"str":   "hello",
		"intr":  2.0,
		"float": 55.5,
		"bool":  true,
		"obj": map[string]interface{}{
			"nest": map[string]interface{}{
				"test": "hello",
			},
			"hello": "world",
		},
		"arr":   []interface{}{"test", 123.0},
		"empty": nil,
	}
	return jsonMsg
}

// This function publishes the message to subscriber checks whether the plugin has processed it ot not
func testSubscriber(t *testing.T, eisMsgBus *EisMsgbus, jsonMsg map[string]interface{}, testCaseName string) {
	err := eisMsgBus.Start(eisMsgBus.ac)
	assert.NoError(t, err)

	buffer, err := json.Marshal(jsonMsg)
	buffer = buffer
	assert.NoError(t, err)

	for key, _ := range eisMsgBus.pluginConfigObj.mapOfPrefixToConfig {
		msg := eismsgbustype.NewMsgEnvelope(nil, buffer)
		msg.Name = key
		for _, sub := range eisMsgBus.pluginSubObj.msgBusSubMap {
			sub.MessageChannel <- msg
		}
	}

	time.Sleep(10000 * time.Millisecond)

	for _, sub := range eisMsgBus.pluginSubObj.msgBusSubMap {
		numElm := len(sub.MessageChannel)
		assert := assert.New(t)
		message := testCaseName + " failed"
		assert.Equal(numElm, 0, message)
	}

	go eisMsgBus.Stop()
}

// Test With Smaple publisher
func TestSubscriber(t *testing.T) {
	fmt.Printf("\n===========In TestSubscriber=========\n")
	eisMsgBus := NewTestEisMsgbus()
	jsonMsg := getTestJsonData()
	testSubscriber(t, eisMsgBus, jsonMsg, "TestSubscriber")
}

func TestSubscriberWithProfiler(t *testing.T) {
	fmt.Printf("\n===========In TestSubscriberWithProfiler=========\n")
	eisMsgBus := NewTestEisMsgbus()
	eisMsgBus.pluginConfigObj.profiling = true
	jsonMsg := getTestJsonData()
	testSubscriber(t, eisMsgBus, jsonMsg, "TestSubscriberWithProfiler")
}
