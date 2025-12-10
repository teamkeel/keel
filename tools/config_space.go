package tools

import toolsproto "github.com/teamkeel/keel/tools/proto"

type SpaceConfigs []*SpaceConfig

type SpaceConfig struct {
	ID           string        `json:"id"`
	Name         string        `json:"name"`
	Icon         string        `json:"icon"`
	DisplayOrder int32         `json:"display_order"`
	Actions      LinkConfigs   `json:"actions"`
	Groups       SpaceGroups   `json:"groups"`
	Links        ExternalLinks `json:"links"`
	Metrics      SpaceMetrics  `json:"metrics"`
}

// toProto will return the SpaceConfigs as protobuf messages.
func (s SpaceConfigs) toProto() []*toolsproto.Space {
	spaces := []*toolsproto.Space{}
	for _, cfg := range s {
		spaces = append(spaces, &toolsproto.Space{
			Id:           cfg.ID,
			Name:         cfg.Name,
			Icon:         cfg.Icon,
			DisplayOrder: cfg.DisplayOrder,
			Actions:      cfg.Actions.applyOn(nil),
			Links:        cfg.Links.toProto(),
			Groups:       cfg.Groups.toProto(),
			Metrics:      cfg.Metrics.toProto(),
		})
	}

	return spaces
}

type SpaceGroups []*SpaceGroup

type SpaceGroup struct {
	ID           string      `json:"id"`
	Name         string      `json:"name"`
	Description  string      `json:"description"`
	DisplayOrder int32       `json:"display_order"`
	Actions      LinkConfigs `json:"actions"`
}

func (g SpaceGroups) toProto() []*toolsproto.SpaceGroup {
	groups := []*toolsproto.SpaceGroup{}
	for _, cfg := range g {
		groups = append(groups, &toolsproto.SpaceGroup{
			Id:           cfg.ID,
			Name:         makeStringTemplate(&cfg.Name),
			Description:  makeStringTemplate(&cfg.Description),
			DisplayOrder: cfg.DisplayOrder,
			Actions:      cfg.Actions.applyOn(nil),
		})
	}

	return groups
}

type SpaceMetrics []*SpaceMetric

type SpaceMetric struct {
	Label         string `json:"label"`
	ToolID        string `json:"tool_id"`
	FacetLocation string `json:"facet_location"`
	DisplayOrder  int32  `json:"display_order"`
}

func (m SpaceMetrics) toProto() []*toolsproto.SpaceMetric {
	metrics := []*toolsproto.SpaceMetric{}
	for _, cfg := range m {
		metrics = append(metrics, &toolsproto.SpaceMetric{
			Label:         makeStringTemplate(&cfg.Label),
			ToolId:        cfg.ToolID,
			FacetLocation: &toolsproto.JsonPath{Path: cfg.FacetLocation},
			DisplayOrder:  cfg.DisplayOrder,
		})
	}

	return metrics
}
