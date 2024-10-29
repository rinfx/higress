package wrapper

import (
	"encoding/json"
	"fmt"

	"github.com/higress-group/proxy-wasm-go-sdk/proxywasm"
	"github.com/tidwall/gjson"
)

const CustomLogKey = "custom_log"

func UnmarshalStr(marshaledJsonStr string) string {
	// e.g. "{\"feild1\":\"value1\",\"feild2\":\"value2\"}"
	var jsonStr string
	err := json.Unmarshal([]byte(marshaledJsonStr), &jsonStr)
	if err != nil {
		proxywasm.LogErrorf("failed to unmarshal json string, raw string is: %s", marshaledJsonStr)
		return ""
	}
	// e.g. {"feild1":"value1","feild2":"value2"}
	return jsonStr
}

func MarshalStr(raw string) string {
	// e.g. {"feild1":"value1","feild2":"value2"}
	helper := map[string]string{
		"placeholder": raw,
	}
	marshaledHelper, _ := json.Marshal(helper)
	marshaledRaw := gjson.GetBytes(marshaledHelper, "placeholder").Raw
	if len(marshaledRaw) > 2 {
		// e.g. {\"feild1\":\"value1\",\"feild2\":\"value2\"}
		return marshaledRaw[1 : len(marshaledRaw)-1]
	} else {
		proxywasm.LogErrorf("failed to marshal json string, raw string is: %s", raw)
		return ""
	}
}

func ExtendAccessLog(items map[string]interface{}) error {
	// e.g. {\"feild1\":\"value1\",\"feild2\":\"value2\"}
	preMarshaledJsonLogStr, _ := proxywasm.GetProperty([]string{CustomLogKey})
	customLog := map[string]interface{}{}
	if string(preMarshaledJsonLogStr) != "" {
		// e.g. {"feild1":"value1","feild2":"value2"}
		preJsonLogStr := UnmarshalStr(fmt.Sprintf(`"%s"`, string(preMarshaledJsonLogStr)))
		err := json.Unmarshal([]byte(preJsonLogStr), &customLog)
		if err != nil {
			proxywasm.LogErrorf("failed to unmarshal custom_log, will overwrite old custom_log")
		}
	}
	// update customLog
	for k, v := range items {
		customLog[k] = v
	}
	// e.g. {"feild1":"value1","feild2":2,"field3":"value3"}
	jsonStr, _ := json.Marshal(customLog)
	// e.g. {\"feild1\":\"value1\",\"feild2\":2,\"field3\":\"value3\"}
	marshaledJsonStr := MarshalStr(string(jsonStr))
	if err := proxywasm.SetProperty([]string{CustomLogKey}, []byte(marshaledJsonStr)); err != nil {
		proxywasm.LogErrorf("failed to set custom_log in filter state, raw is %s", marshaledJsonStr)
		return err
	}
	return nil
}
