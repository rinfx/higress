package main

import (
	"fmt"

	"github.com/alibaba/higress/plugins/wasm-go/pkg/wrapper"
	"github.com/higress-group/proxy-wasm-go-sdk/proxywasm"
	"github.com/higress-group/proxy-wasm-go-sdk/proxywasm/types"
	"github.com/tidwall/gjson"
)

func main() {
	wrapper.SetCtx(
		"ding",
		wrapper.ParseConfigBy(parseConfig),
		wrapper.ProcessRequestHeadersBy(onHttpRequestHeaders),
		wrapper.ProcessResponseHeadersBy(onHttpResponseHeaders),
		wrapper.ProcessRequestBodyBy(onHttpRequestBody),
		wrapper.ProcessResponseBodyBy(onHttpResponseBody),
	)
}

type DingConfig struct{}

const defaultRequestFormat = `{"model": "gpt-3.5-turbo","messages": [{"role": "user","content": "%s"}]}`
const defaultResponseFormat = `{"msgtype":"markdown","markdown": {"title": "answer", "text": "%s"}}`

func parseConfig(json gjson.Result, config *DingConfig, log wrapper.Log) error {
	return nil
}

func onHttpRequestHeaders(ctx wrapper.HttpContext, config DingConfig, log wrapper.Log) types.Action {
	proxywasm.RemoveHttpRequestHeader("content-length")
	return types.ActionContinue
}

func onHttpRequestBody(ctx wrapper.HttpContext, config DingConfig, body []byte, log wrapper.Log) types.Action {
	question := gjson.GetBytes(body, "text.content").String()
	if question == "" {
		log.Error("Parse request body error")
	}
	newBody := fmt.Sprintf(defaultRequestFormat, question)
	proxywasm.ReplaceHttpRequestBody([]byte(newBody))
	return types.ActionContinue
}

func onHttpResponseHeaders(ctx wrapper.HttpContext, config DingConfig, log wrapper.Log) types.Action {
	proxywasm.RemoveHttpResponseHeader("content-length")
	return types.ActionContinue
}

func onHttpResponseBody(ctx wrapper.HttpContext, config DingConfig, body []byte, log wrapper.Log) types.Action {
	answer := gjson.GetBytes(body, "choices.0.message.content").String()
	if answer == "" {
		log.Error("Parse response body error")
	}
	newBody := fmt.Sprintf(defaultResponseFormat, answer)
	proxywasm.ReplaceHttpResponseBody([]byte(newBody))
	return types.ActionContinue
}
