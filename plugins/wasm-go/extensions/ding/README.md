# 介绍
提供AI可观测基础能力，其后需接ai-proxy插件，如果不接ai-proxy插件的话，则只支持openai协议。

# 配置说明

| 名称         | 数据类型   | 填写要求 | 默认值 | 描述               |
|------------|--------|------|-----|------------------|
| `enable` | bool | 必填   | -   | 是否开启ai统计功能 |

开启后能够支持产生网关粒度、路由粒度、服务粒度、模型粒度的metric以及log

metrics 示例：
```
route_upstream_model_input_token{ai_route="openai",ai_cluster="qwen",ai_model="qwen-max"} 21
route_upstream_model_output_token{ai_route="openai",ai_cluster="qwen",ai_model="qwen-max"} 17
```

log 示例：

```json
{
    "model": "qwen-max",
    "input_token": "21",
    "output_token": "17",
    ...
}
```