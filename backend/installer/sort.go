package installer

import (
	"bytes"
	"encoding/json"
	"sort"
)

// marshalSorted 输出 key 字典序的 JSON, 缩进 2 空格;
// 让多次写入 settings.json 的 diff 稳定, 也方便用户人眼检查
func marshalSorted(m map[string]json.RawMessage) ([]byte, error) {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var buf bytes.Buffer
	buf.WriteString("{\n")
	for i, k := range keys {
		kj, err := json.Marshal(k)
		if err != nil {
			return nil, err
		}
		var pretty bytes.Buffer
		if err := json.Indent(&pretty, m[k], "  ", "  "); err != nil {
			pretty.Write(m[k])
		}
		buf.WriteString("  ")
		buf.Write(kj)
		buf.WriteString(": ")
		buf.Write(pretty.Bytes())
		if i < len(keys)-1 {
			buf.WriteString(",")
		}
		buf.WriteString("\n")
	}
	buf.WriteString("}\n")
	return buf.Bytes(), nil
}
