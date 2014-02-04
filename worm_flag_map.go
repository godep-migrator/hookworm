package hookworm

import (
	"encoding/json"
	"fmt"
	"strings"
)

type wormFlagMap struct {
	values map[string]interface{}
}

func newWormFlagMap() *wormFlagMap {
	return &wormFlagMap{
		values: make(map[string]interface{}),
	}
}

func (wfm *wormFlagMap) String() string {
	s := ""
	for k, v := range wfm.values {
		s += fmt.Sprintf("%s=%v;", k, v)
	}
	return s
}

func (wfm *wormFlagMap) Get(key string) interface{} {
	if value, ok := wfm.values[key]; ok {
		return value
	}
	return ""
}

func (wfm *wormFlagMap) Set(value string) error {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil
	}

	pairs := strings.Split(value, ";")

	for _, pair := range pairs {
		parts := strings.SplitN(pair, "=", 2)
		if len(parts) == 2 {
			k := parts[0]
			v := parts[1]
			switch strings.ToLower(v) {
			case "true", "yes", "on":
				wfm.values[k] = true
			case "false", "no", "off":
				wfm.values[k] = false
			default:
				wfm.values[k] = strings.Trim(v, "\"' ")
			}
		} else {
			wfm.values[parts[0]] = true
		}
	}

	return nil
}

func (wfm *wormFlagMap) MarshalJSON() ([]byte, error) {
	return json.Marshal(wfm.values)
}

func (wfm *wormFlagMap) UnmarshalJSON(raw []byte) error {
	return json.Unmarshal(raw, &wfm.values)
}
