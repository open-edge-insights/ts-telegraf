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
	"fmt"
	"github.com/influxdata/telegraf"
	agent "github.com/influxdata/telegraf/agent"
	jsonParser "github.com/influxdata/telegraf/plugins/parsers/json"
	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/assert"
	"testing"
)

type testMetricMaker struct {
	logger telegraf.Logger
}

func (_ *testMetricMaker) LogName() string {
	return "testMetricMaker"
}

func (_ *testMetricMaker) MakeMetric(metric telegraf.Metric) telegraf.Metric {
	return metric
}

func (tmMaket *testMetricMaker) Log() telegraf.Logger {
	return tmMaket.logger
}

func NewTestEiiMsgbus() *EiiMsgbus {
	parserConfig := jsonParser.Config{
		MetricName: "sample1",
		Strict:     true,
	}
	parser, err := jsonParser.New(&parserConfig)
	if err != nil {
		fmt.Printf("error while creating the parse object %v", err)
	}

	var tmMake agent.MetricMaker
	tmMake = &testMetricMaker{logger: testutil.Logger{}}
	//acc := agent.NewAccumulator(tmMake, make(telegraf.Metric))
	temp := EiiMsgbus{
		Log:           testutil.Logger{},
		InstanceName: "publisher1",
		parser:        parser,
		ac:            agent.NewAccumulator(tmMake, make(chan telegraf.Metric, 100)),
	}

	return &temp
}

// Test Plugins start and stop
func TestStart(t *testing.T) {
	fmt.Printf("\n===========In TestStart=========\n")
	eiiMsgBus := NewTestEiiMsgbus()
	var acc telegraf.Accumulator
	err := eiiMsgBus.Start(acc)
	assert.NoError(t, err)
	eiiMsgBus.Stop()
}

// Test Gather
func TestGather(t *testing.T) {
	fmt.Printf("\n===========In TestGather=========\n")
	eiiMsgBus := NewTestEiiMsgbus()
	err := eiiMsgBus.Gather(nil)
	assert.NoError(t, err)
}

// Test SampleConfig
func TestSampleConfig(t *testing.T) {
	fmt.Printf("\n===========In SampleConfig=========\n")
	eiiMsgBus := NewTestEiiMsgbus()
	config := eiiMsgBus.SampleConfig()
	assert.NotEmpty(t, config)
}
