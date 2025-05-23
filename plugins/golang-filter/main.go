package main

import (
	"net/http"

	llm_scheduler "github.com/alibaba/higress/plugins/golang-filter/llm-scheduler"
	mcp_server "github.com/alibaba/higress/plugins/golang-filter/mcp-server"
	mcp_session "github.com/alibaba/higress/plugins/golang-filter/mcp-session"
	"github.com/envoyproxy/envoy/contrib/golang/common/go/api"
	envoyHttp "github.com/envoyproxy/envoy/contrib/golang/filters/http/source/go/pkg/http"
)

func init() {
	envoyHttp.RegisterHttpFilterFactoryAndConfigParser(mcp_session.Name, mcp_session.FilterFactory, &mcp_session.Parser{})
	envoyHttp.RegisterHttpFilterFactoryAndConfigParser(mcp_server.Name, mcp_server.FilterFactory, &mcp_server.Parser{})
	envoyHttp.RegisterHttpFilterFactoryAndConfigParser(llm_scheduler.Name, llm_scheduler.FilterFactory, &llm_scheduler.Parser{})
	go func() {
		defer func() {
			if r := recover(); r != nil {
				api.LogErrorf("PProf server recovered from panic: %v", r)
			}
		}()
		api.LogError(http.ListenAndServe("localhost:6060", nil).Error())
	}()
}

func main() {}
