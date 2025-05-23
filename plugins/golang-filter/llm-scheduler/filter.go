package llm_scheduler

import (
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"strings"

	"github.com/alibaba/higress/plugins/golang-filter/llm-scheduler/backend"
	"github.com/alibaba/higress/plugins/golang-filter/llm-scheduler/backend/vllm"
	"github.com/alibaba/higress/plugins/golang-filter/llm-scheduler/scheduling"

	"github.com/envoyproxy/envoy/contrib/golang/common/go/api"
	"github.com/prometheus/common/expfmt"
)

type filter struct {
	api.PassThroughStreamFilter

	config    *config
	buffer    string
	scheduler *scheduling.Scheduler
	callbacks api.FilterCallbackHandler
	// path      string
	// reqHeader api.RequestHeaderMap
}

// Callbacks which are called in request path
func (f *filter) DecodeHeaders(header api.RequestHeaderMap, endStream bool) api.StatusType {
	if header.Method() != "POST" {
		return api.Continue
	}
	return api.StopAndBuffer
}

func (f *filter) DecodeData(buffer api.BufferInstance, endStream bool) api.StatusType {
	f.buffer += buffer.String()
	if !endStream {
		return api.StopAndBuffer
	}
	if err := f.SetScheduler(); err != nil {
		api.LogDebugf("Set scheduler failed: %v", err)
		return api.Continue
	}
	var rb map[string]interface{}
	if err := json.Unmarshal([]byte(f.buffer), &rb); err != nil {
		return api.Continue
	}
	model, ok := rb["model"].(string)
	if !ok {
		api.LogDebug("model not found in request")
		return api.Continue
	}
	llmReq := &scheduling.LLMRequest{
		Model:    model,
		Critical: f.config.isCritical(model),
	}
	api.LogDebugf("config: %+v, llmReq: %+v", f.config, llmReq)
	targetPod, err := f.scheduler.Schedule(llmReq)
	if err != nil {
		api.LogDebugf("failed to find target pod: %v", err)
		f.callbacks.DecoderFilterCallbacks().SendLocalReply(http.StatusTooManyRequests, "Request is dropped due to limited backend resources", nil, 0, "")
		return api.LocalReply
	}
	if isValidAddress(targetPod.Address) {
		api.LogDebugf("override upstream host: %s", targetPod.Address)
		f.callbacks.DecoderFilterCallbacks().SetOverrideUpstreamHost(targetPod.Address)
	} else {
		api.LogDebugf("override upstream host failed, address is invalid: %s", targetPod.Address)
	}
	return api.Continue
}

func (f *filter) SetScheduler() error {
	infos := f.callbacks.StreamInfo().DynamicMetadata().Get("envoy.filters.http.ip_setting")
	if len(infos) == 0 {
		return errors.New("backend is not support llm scheduling")
	}
	// f.callbacks.Log(api.Info, fmt.Sprintf("%+v", infos))
	var pms []*backend.PodMetrics
	for addr, metric := range infos {
		parser := expfmt.TextParser{}
		metricFamilies, err := parser.TextToMetricFamilies(strings.NewReader(metric.(string)))
		if err != nil {
			api.LogError(fmt.Sprint(err))
			return err
		}
		// api.LogInfof("address: %s", addr)
		// f.callbacks.Log(api.Info, fmt.Sprintf("metrics: %+v", metricFamilies))
		// for k, v := range metricFamilies {
		// 	fmt.Printf("---------------------------------------\nkey: %s, value: %+v\n", k, v)
		// }
		pm := &backend.PodMetrics{
			Pod: backend.Pod{
				Name:    addr,
				Address: addr,
			},
			Metrics: backend.Metrics{},
		}
		pm, err = vllm.PromToPodMetrics(metricFamilies, pm)
		if err != nil {
			api.LogError(fmt.Sprint(err))
			return err
		}
		pms = append(pms, pm)
	}
	f.scheduler = scheduling.NewScheduler(pms)
	return nil
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
