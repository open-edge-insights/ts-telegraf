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

package eii_msgbus

import (
	eiimsgbustype "EIIMessageBus/pkg/types"
	"encoding/json"
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

func NewTopicRuntimeData() *tpRuntimeData {
	temp := new(tpRuntimeData)
	temp.mName = "temperature"
	temp.tpName = "temperature"
	return temp
}

func TestSimpleMsgProcessor(t *testing.T) {
	fmt.Printf("\n===========In SimpleMsgProcessor=========\n")
	eiiMsgBus := NewTestEiiMsgbus()
	var processor simpleMsgProcessor
	topicRtData := NewTopicRuntimeData()
	topicRtData.parser = eiiMsgBus.parser
	topicRtData.writer = telegrafAccWriter{ac: eiiMsgBus.ac}
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

	msg := eiimsgbustype.NewMsgEnvelope(nil, buffer)
	msg.Name = "topic-name"
	err = processor.processData(topicRtData, dataFromMsgBus{msg: msg, profInfo: nil})
	assert.NoError(t, err)
}
