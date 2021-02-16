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

package main

import (
	eiimsgbus "EIIMessageBus/eiimsgbus"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strconv"
	"time"
)

func main() {
	configFile := flag.String("configFile", "", "JSON configuration file")
	topic := flag.String("topic", "", "Subscription topic")
	count := flag.String("count", "", "number of messages")
	interval := flag.String("interval", "", "number of messages")
	flag.Parse()

	if *configFile == "" {
		fmt.Println("-- Config file must be specified")
		return
	}

	fmt.Printf("-- Loading configuration file %s\n", *configFile)
	config, err := eiimsgbus.ReadJsonConfig(*configFile)
	if err != nil {
		fmt.Printf("-- Failed to parse config: %v\n", err)
		return
	}

	fmt.Println("-- Initializing message bus context")
	client, err := eiimsgbus.NewMsgbusClient(config)
	if err != nil {
		fmt.Printf("-- Error initializing message bus context: %v\n", err)
		return
	}
	defer client.Close()

	fmt.Printf("-- Creating publisher for topic %s\n", *topic)
	publisher, err := client.NewPublisher(*topic)
	if err != nil {
		fmt.Printf("-- Error creating publisher: %v\n", err)
		return
	}
	defer publisher.Close()

	cnt, err := strconv.ParseInt(*count, 10, 32)
	if err != nil {
		fmt.Printf("-- Error in strconv.ParseInt: %v\n", err)
		return
	}

	intval, err := time.ParseDuration(*interval)
	if err != nil {
		fmt.Printf("-- Error in time.ParseDuration: %v\n", err)
		return
	}

	fmt.Println("-- Running...")
	msg := map[string]interface{}{
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

	buffer, err := json.Marshal(msg)

	if err != nil {
		fmt.Printf("-- Failed to Marshal teh message : %v\n", err)
		os.Exit(0)
	}

	for i := int64(0); i < cnt; i++ {
		err = publisher.Publish(buffer)
		if err != nil {
			fmt.Printf("-- Failed to publish message: %v\n", err)
			return
		}
		time.Sleep(intval)
	}
}
