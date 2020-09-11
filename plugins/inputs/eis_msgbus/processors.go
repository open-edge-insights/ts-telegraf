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
	"fmt"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/parsers"
	"time"
)

// Does JSON parsing using telegraf JSON parser
func doJSONParsing(parser parsers.Parser, data dataFromMsgBus) (*[]telegraf.Metric, error) {
	var t1 int64

	if data.profInfo != nil {
		t1 = time.Now().UnixNano()
	}

	metrics, err := parser.Parse(data.msg.Blob)

	if data.profInfo != nil {
		data.profInfo["total_time_spent_in_json_parser"] = time.Now().UnixNano() - t1
	}

	if err != nil {
		return nil, fmt.Errorf("Error in json parsing:%v", err)
	}

	return &metrics, nil
}

// Sets the measurement name in every metric and writes it to telegraf
func (writer *telegrafAccWriter) writeToTelegraf(metrics *[]telegraf.Metric, mName string, profInfo map[string]interface{}) {

	for _, elm := range *metrics {
		elm.SetName(mName)

		if profInfo != nil {
			profInfo["total_time_spent_in_plugin"] = time.Now().UnixNano() - profInfo["ts_plugin_in"].(int64)
			delete(profInfo, "ts_plugin_in")
			for key, value := range profInfo {
				elm.AddField(key, value)
			}
		}
		writer.ac.AddMetric(elm)
	}
}

// Simple processor which parses the json data and push the metrics into telegraf engine
func (processor simpleMsgProcessor) processData(tpRtInfo *tpRuntimeData, data dataFromMsgBus) error {
	metrics, err := doJSONParsing(tpRtInfo.parser, data)
	if err != nil {
		return err
	}
	tpRtInfo.writer.writeToTelegraf(metrics, tpRtInfo.mName, data.profInfo)
	return nil
}
