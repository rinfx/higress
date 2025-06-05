package main

import (
	"errors"
	"net"
	"strings"

	"ai-load-balancer/backend"
	"ai-load-balancer/backend/vllm"
	"ai-load-balancer/scheduling"

	"github.com/alibaba/higress/plugins/wasm-go/pkg/wrapper"
	"github.com/higress-group/proxy-wasm-go-sdk/proxywasm"
	"github.com/higress-group/proxy-wasm-go-sdk/proxywasm/types"
	"github.com/prometheus/common/expfmt"
	"github.com/tidwall/gjson"
)

func main() {}

func init() {
	wrapper.SetCtx(
		"ai-load-balancer",
		wrapper.ParseConfigBy(parseConfig),
		wrapper.ProcessRequestHeadersBy(onHttpRequestHeaders),
		wrapper.ProcessRequestBodyBy(onHttpRequestBody),
	)
}

type LBConfig struct {
	key    string
	count  int64
	window int64
}

func parseConfig(json gjson.Result, config *LBConfig, log wrapper.Log) error {
	return nil
}

// Callbacks which are called in request path
func onHttpRequestHeaders(ctx wrapper.HttpContext, config LBConfig, log wrapper.Log) types.Action {
	return types.ActionContinue
}

func onHttpRequestBody(ctx wrapper.HttpContext, config LBConfig, body []byte, log wrapper.Log) types.Action {
	llmReq := &scheduling.LLMRequest{
		Model:    "meta-llama/Llama-2-7b-hf",
		Critical: true,
	}
	hostInfos, err := proxywasm.GetUpstreamHosts()
	log.Infof("%+v", hostInfos)
	if err != nil {
		return types.ActionContinue
	}
	hostMetrics := make(map[string]string)
	for _, hostInfo := range hostInfos {
		hostMetrics[hostInfo[0]] = gjson.Get(hostInfo[1], "metrics").String()
	}
	// log.Infof("%v", hostMetrics)
	// for addr, metric := range hostMetrics {
	// 	log.Infof("--------\naddress: %s\nmetrics:\n%s", addr, metric)
	// }
	scheduler, err := GetScheduler(hostMetrics)
	if err != nil {
		log.Infof("initial scheduler failed: %v", err)
		return types.ActionContinue
	}
	targetPod, err := scheduler.Schedule(llmReq)
	if err != nil {
		log.Infof("pod select failed: %v", err)
		return types.ActionContinue
	}
	if isValidAddress(targetPod.Address) {
		proxywasm.SetUpstreamOverrideHost([]byte(targetPod.Address))
	}
	return types.ActionContinue
}

func GetScheduler(hostMetrics map[string]string) (*scheduling.Scheduler, error) {
	if len(hostMetrics) == 0 {
		return nil, errors.New("backend is not support llm scheduling")
	}
	var pms []*backend.PodMetrics
	for addr, metric := range hostMetrics {
		parser := expfmt.TextParser{}
		metricFamilies, err := parser.TextToMetricFamilies(strings.NewReader(metric))
		if err != nil {
			return nil, err
		}
		pm := &backend.PodMetrics{
			Pod: backend.Pod{
				Name:    addr,
				Address: addr,
			},
			Metrics: backend.Metrics{},
		}
		pm, err = vllm.PromToPodMetrics(metricFamilies, pm)
		if err != nil {
			return nil, err
		}
		pms = append(pms, pm)
	}
	return scheduling.NewScheduler(pms), nil
}

func isValidAddress(s string) bool {
	host, port, err := net.SplitHostPort(s)
	if err != nil {
		return false
	}

	_, err = net.LookupPort("tcp", port)
	if err != nil {
		return false
	}

	ip := net.ParseIP(host)
	return ip != nil
}
