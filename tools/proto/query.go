package proto

import (
	"encoding/json"

	"google.golang.org/protobuf/encoding/protojson"
)

// ActionConfigs will return all the actionConfigs in this selection of tools
func (tools *Tools) ActionConfigs() []*ActionConfig {
	if tools == nil {
		return nil
	}
	cfgs := []*ActionConfig{}
	for _, t := range tools.Configs {
		if t.Type == Tool_ACTION {
			cfgs = append(cfgs, t.ActionConfig)
		}
	}

	return cfgs
}

// FindByID finds a tool in the given tools message by id
func (tools *Tools) FindByID(id string) *Tool {
	if tools == nil {
		return nil
	}

	for _, t := range tools.Configs {
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

// GetOperationName will return the name of the operation that drives this tool.
//
// For action based tools, this will be the actionName, for flow based tools, the flow name
func (t *Tool) GetOperationName() string {
	if t.IsActionBased() {
		return t.GetActionConfig().ActionName
	}

	return t.GetFlowConfig().FlowName
}

// IsActionBased checks if the tool is driven by an API action
func (t *Tool) IsActionBased() bool {
	return t.Type == Tool_ACTION && t.ActionConfig != nil
}

// ToTool transforms this ActionConfig into an action based Tool wrapper message.
func (t *ActionConfig) ToTool() *Tool {
	if t == nil {
		return nil
	}

	return &Tool{
		Id:           t.Id,
		Type:         Tool_ACTION,
		ActionConfig: t,
	}
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

func FindLinkByToolID(links []*ToolLink, toolID string) *ToolLink {
	for _, l := range links {
		if l.ToolId == toolID {
			return l
		}
	}
	return nil
}

func FindToolGroupLinkByToolID(links []*ToolGroup_GroupActionLink, toolID string) *ToolGroup_GroupActionLink {
	for _, l := range links {
		if l.ActionLink.ToolId == toolID {
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

func (l *ToolLink) GetJSONDataMapping() string {
	dm := l.GetObjDataMapping()
	str, err := json.Marshal(dm)
	if err != nil {
		return ""
	}
	return string(str)
}

func (l *ToolLink) GetObjDataMapping() []any {
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

func (l *ToolGroup_GroupActionLink) GetResponseOverridesMap() map[string]bool {
	m := map[string]bool{}
	for _, ro := range l.GetResponseOverrides() {
		m[ro.GetFieldLocation().GetPath()] = ro.GetVisible()
	}
	return m
}

func (dl *DisplayLayoutConfig) AllToolLinks() []*ToolLink {
	links := []*ToolLink{}
	if dl == nil {
		return nil
	}

	if dl.GetInboxConfig() != nil {
		if dl.GetInboxConfig().GetTool != nil {
			links = append(links, dl.GetInboxConfig().GetTool)
		}
	}
	if dl.GetBoardConfig() != nil {
		if dl.GetBoardConfig().GetTool != nil {
			links = append(links, dl.GetBoardConfig().GetTool)
		}
		if dl.GetBoardConfig().UpdateAction != nil {
			links = append(links, dl.GetBoardConfig().UpdateAction)
		}
	}

	if dl.GetGridConfig() != nil {
		if dl.GetGridConfig().UpdateAction != nil {
			links = append(links, dl.GetGridConfig().UpdateAction)
		}
	}

	return links
}
