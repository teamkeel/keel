package proto

import (
	"encoding/json"

	"google.golang.org/protobuf/encoding/protojson"
)

// FindByID finds a tool in the given tools message by id
func (tools *Tools) FindByID(id string) *ActionConfig {
	if tools == nil {
		return nil
	}

	for _, t := range tools.Tools {
		if t.Id == id {
			return t
		}
	}

	return nil
}

// DiffIDs will return a subset of the given tools which do not exist in our current tools wrapper
func (tools *Tools) DiffIDs(ids []string) []string {
	diffs := []string{}
	for _, id := range ids {
		if tools.FindByID(id) == nil {
			diffs = append(diffs, id)
		}
	}

	return diffs
}

func (t *ActionConfig) FindInput(location *JsonPath) *RequestFieldConfig {
	for _, f := range t.GetInputs() {
		if f.GetFieldLocation().GetPath() == location.GetPath() {
			return f
		}
	}

	return nil
}

func (t *ActionConfig) FindInputByPath(location string) *RequestFieldConfig {
	for _, f := range t.GetInputs() {
		if f.GetFieldLocation().GetPath() == location {
			return f
		}
	}

	return nil
}

func (t *ActionConfig) FindResponse(location *JsonPath) *ResponseFieldConfig {
	for _, f := range t.GetResponse() {
		if f.GetFieldLocation().GetPath() == location.GetPath() {
			return f
		}
	}

	return nil
}

func (t *ActionConfig) FindResponseByPath(location string) *ResponseFieldConfig {
	for _, f := range t.GetResponse() {
		if f.GetFieldLocation().GetPath() == location {
			return f
		}
	}

	return nil
}

func (c *Capabilities) Diff(other *Capabilities) map[string]bool {
	diffs := map[string]bool{}
	if c.GetAudit() != other.GetAudit() {
		diffs["audit"] = other.GetAudit()
	}
	if c.GetComments() != other.GetComments() {
		diffs["comments"] = other.GetComments()
	}
	return diffs
}

func FindLinkByToolID(links []*ActionLink, toolID string) *ActionLink {
	for _, l := range links {
		if l.ToolId == toolID {
			return l
		}
	}
	return nil
}

func FindToolGroupByID(groups []*ToolGroup, id string) *ToolGroup {
	for _, g := range groups {
		if g.Id == id {
			return g
		}
	}
	return nil
}

func (l *ActionLink) GetJSONDataMapping() string {
	dm := l.GetObjDataMapping()
	str, err := json.Marshal(dm)
	if err != nil {
		return ""
	}
	return string(str)
}

func (l *ActionLink) GetObjDataMapping() []any {
	ret := []any{}
	for _, d := range l.GetData() {
		jsData, err := protojson.Marshal(d)
		if err != nil {
			continue
		}
		var datum any
		if err = json.Unmarshal(jsData, &datum); err != nil {
			continue
		}
		ret = append(ret, datum)
	}
	return ret
}

func (dl *DisplayLayoutConfig) JSON() string {
	if dl == nil {
		return ""
	}

	str, err := json.Marshal(dl)
	if err != nil {
		return ""
	}
	return string(str)
}

func (dl *DisplayLayoutConfig) AsObj() any {
	if dl == nil {
		return nil
	}
	str := dl.JSON()
	var d any
	if err := json.Unmarshal([]byte(str), &d); err != nil {
		return nil
	}

	return d
}
