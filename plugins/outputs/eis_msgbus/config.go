
/*
Copyright (c) 2021 Intel Corporation.
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
	"strconv"
)

// Converts the plugin configuration into plugin configuration object of type pluginConfigObj.
func (pluginConfigObj *eisMsgbusOutputPluginConfig) initConfig(emb *EisMsgbus) error {

	appConfig, err := emb.confMgr.GetAppConfig()
	if err != nil {
		return fmt.Errorf("Error in getting config: %v", err)
	}

	var instanceConfig map[string]interface{}
	var found bool
	if instanceConfig, found = appConfig[emb.Instance_name].(map[string]interface{}); found == false {
		return fmt.Errorf("Could not get the configuration for %v: %v", emb.Instance_name, err)
	}

	value := instanceConfig["profiling"].(string)
	pluginConfigObj.profiling, err = strconv.ParseBool(value)
	if err != nil {
		return fmt.Errorf("Parsing profiling mode failed: %v", err)
	}

	tempMeasurements := instanceConfig["measurements"].([]interface{})
	pluginConfigObj.measurements = make([]string, len(tempMeasurements))
	for idx, measurement := range tempMeasurements {
		pluginConfigObj.measurements[idx] = fmt.Sprint(measurement)
	}

	devMode, err := emb.confMgr.IsDevMode()
	pluginConfigObj.devmode = devMode
	if err != nil {
		return fmt.Errorf("Fail to read DEV_MODE from etcd: %v", err)
	}

	pluginConfigObj.instanceName = emb.Instance_name

	return nil
}

