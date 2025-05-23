package llm_scheduler

import (
	xds "github.com/cncf/xds/go/xds/type/v3"
	"google.golang.org/protobuf/types/known/anypb"

	"github.com/envoyproxy/envoy/contrib/golang/common/go/api"
)

const Name = "llm-scheduler"

type config struct {
	criticalModels map[string]struct{}
}

func (c *config) isCritical(model string) bool {
	_, exists := c.criticalModels[model]
	return exists
}

type Parser struct {
}

func (p *Parser) Parse(any *anypb.Any, callbacks api.ConfigCallbackHandler) (interface{}, error) {
	configStruct := &xds.TypedStruct{}
	if err := any.UnmarshalTo(configStruct); err != nil {
		return nil, err
	}
	v := configStruct.Value

	conf := &config{}
	if criticalModels, ok := v.AsMap()["criticalModels"].([]interface{}); ok {
		conf.criticalModels = make(map[string]struct{})
		for _, modelRaw := range criticalModels {
			if model, ok := modelRaw.(string); ok {
				conf.criticalModels[model] = struct{}{}
			}
		}
	} else {
		api.LogErrorf("parse configuration failed, raw configuration is: %+v", v.String())
	}

	return conf, nil
}

func (p *Parser) Merge(parent interface{}, child interface{}) interface{} {
	parentConfig := parent.(*config)
	childConfig := child.(*config)

	// copy one, do not update parentConfig directly.
	newConfig := *parentConfig
	if childConfig.criticalModels != nil {
		newConfig.criticalModels = childConfig.criticalModels
	}
	return &newConfig
}

func FilterFactory(c interface{}, callbacks api.FilterCallbackHandler) api.StreamFilter {
	conf, ok := c.(*config)
	if !ok {
		panic("unexpected config type")
	}
	return &filter{
		callbacks: callbacks,
		config:    conf,
	}
}
