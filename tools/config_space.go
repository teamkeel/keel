package tools

import toolsproto "github.com/teamkeel/keel/tools/proto"

type SpaceConfigs []*SpaceConfig

type SpaceConfig struct {
	ID           string
	Name         string
	Colour       *SpaceColour
	Icon         string
	DisplayOrder int32
	Actions      LinkConfigs
	Groups       SpaceGroups
	Links        ExternalLinks
	Metrics      SpaceMetrics
}

type SpaceColour string

const (
	SpaceColourGreen  SpaceColour = "green"
	SpaceColourBlue   SpaceColour = "blue"
	SpaceColourYellow SpaceColour = "yellow"
	SpaceColourCoral  SpaceColour = "coral"
	SpaceColourPurple SpaceColour = "purple"
)

// toProto will return the SpaceConfigs as protobuf messages.
func (s SpaceConfigs) toProto() []*toolsproto.Space {
	spaces := []*toolsproto.Space{}
	for _, cfg := range s {
		spaces = append(spaces, &toolsproto.Space{
			Id:   cfg.ID,
			Name: cfg.Name,
			Icon: cfg.Icon,
			Colour: func() *toolsproto.Space_Colour {
				if cfg.Colour == nil {
					return nil
				}

				c := toolsproto.Space_Colour(toolsproto.Space_Colour_value[string(*cfg.Colour)])

				return &c
			}(),
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
	ID           string
	Name         string
	Description  string
	DisplayOrder int32
	Actions      LinkConfigs
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
	Label         string
	ToolID        string
	FacetLocation string
	DisplayOrder  int32
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
