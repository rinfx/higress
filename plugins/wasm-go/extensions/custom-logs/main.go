// Copyright (c) 2022 Alibaba Group Holding Ltd.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"github.com/higress-group/proxy-wasm-go-sdk/proxywasm/types"

	"github.com/alibaba/higress/plugins/wasm-go/pkg/wrapper"
)

func main() {
	wrapper.SetCtx(
		"custom-log",
		wrapper.ProcessRequestHeadersBy(onHttpRequestHeaders),
	)
}

type CustomLogConfig struct {
}

func onHttpRequestHeaders(ctx wrapper.HttpContext, config CustomLogConfig, log wrapper.Log) types.Action {
	// log_format: "{\"custom_log\":\"%FILTER_STATE(wasm.custom_log:PLAIN)%\"}"
	object1 := map[string]interface{}{
		"feild1": "value1",
		"feild2": "value2",
	}
	_ = wrapper.ExtendAccessLog(object1)
	object2 := map[string]interface{}{
		"feild2": 1,
		"feild4": "value4",
	}
	_ = wrapper.ExtendAccessLog(object2)
	// log example: {"custom_log":"{\"feild1\":\"value1\",\"feild2\":1,\"feild4\":\"value4\"}"}
	return types.ActionContinue
}
