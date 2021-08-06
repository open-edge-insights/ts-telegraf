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
	"github.com/influxdata/telegraf"
	"time"
)

func (pool *threadPool) initThrPool(processor eiiMsgProcessor, tpRtInfo *tpRuntimeData, poolSize int, Log telegraf.Logger) {
	pool.processor = processor
	pool.tpRtInfo = tpRtInfo
	pool.poolSize = poolSize
	pool.Log = Log
	pool.contrChannel = make(chan byte, poolSize)
}

func (pool *threadPool) setName(name string) {
	pool.name = name
}

func (pool *threadPool) start() {
	for thrID := 0; thrID < pool.poolSize; thrID++ {
		pool.wg.Add(1)
		go pool.thrPoolFunction(thrID)
	}
}

func (pool *threadPool) thrPoolFunction(id int) {
	defer pool.wg.Done()

	pool.Log.Infof("Inside thrPoolFunction for %v with id:%v", pool.tpRtInfo.tpName, id)
	loop := true
	for loop {
		select {
		case msgWrapper := <-pool.tpRtInfo.dataChannel:
			if msgWrapper.profInfo != nil {
				msgWrapper.profInfo["total_time_spent_in_queue"] = time.Now().UnixNano() - msgWrapper.profInfo["ts_queue_in"].(int64)
				delete(msgWrapper.profInfo, "ts_queue_in")
				msgWrapper.profInfo["pool-name"] = pool.name
				msgWrapper.profInfo["thrId"] = id
			}
			err := pool.processor.processData(pool.tpRtInfo, msgWrapper)
			if err != nil {
				pool.Log.Errorf(err.Error())
			}
		case cntrField := <-pool.contrChannel:
			pool.Log.Infof("Exiting(%v) without processing %v elements for topic %v", cntrField, len(pool.tpRtInfo.dataChannel), pool.tpRtInfo.tpName)
			loop = false
		}
	}

	pool.Log.Infof("Exiting theradpool function with id: %v for topic %v", id, pool.tpRtInfo.tpName)
}

func (pool *threadPool) sendShutdownSignal() {
	for thrID := 0; thrID < pool.poolSize; thrID++ {
		//send exit signal to every thread
		pool.contrChannel <- 'E'
	}
}

func (pool *threadPool) waitForShutdown() {
	pool.wg.Wait()
}
