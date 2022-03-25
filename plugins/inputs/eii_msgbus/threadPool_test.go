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
	eiimsgbustype "github.com/open-edge-insights/eii-messagebus-go/pkg/types"
	"encoding/json"
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestThreadPool(t *testing.T) {
	fmt.Printf("\n===========In TestThreadPool=========\n")
	eiiMsgBus := NewTestEiiMsgbus()
	var processor simpleMsgProcessor
	topicRtData := NewTopicRuntimeData()
	topicRtData.parser = eiiMsgBus.parser
	topicRtData.writer = telegrafAccWriter{ac: eiiMsgBus.ac}
	dataChannel := make(chan dataFromMsgBus, 10)
	topicRtData.dataChannel = dataChannel

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

	buffer, err := json.Marshal(jsonMsg)
	assert.NoError(t, err)

	msg := eiimsgbustype.NewMsgEnvelope(nil, buffer)
	msg.Name = "topic-name"
	d := dataFromMsgBus{msg: msg, profInfo: nil}
	dataChannel <- d
	dataChannel <- d
	dataChannel <- d
	dataChannel <- d
	numElm := len(dataChannel)

	pool := threadPool{}
	pool.initThrPool(processor, topicRtData, 2, eiiMsgBus.Log)
	pool.setName("GLOBAL")
	pool.start()
	time.Sleep(5000 * time.Millisecond)
	pool.sendShutdownSignal()
	pool.waitForShutdown()
	numElm = len(dataChannel)
	assert := assert.New(t)
	assert.Equal(numElm, 0, "Test TestThreadPool failed")
}
