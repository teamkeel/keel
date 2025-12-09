package tools

import toolsproto "github.com/teamkeel/keel/tools/proto"

type SpaceConfigs []*SpaceConfig

type SpaceConfig struct {
	ID      string
	Name    string
	Colour  *SpaceColour
	Icon    string
	Actions LinkConfigs
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
			Actions: cfg.Actions.applyOn(nil),
			//TODO: Groups: ,
			//TODO: Links: ,
			//TODO:  Metrics: ,
		})
	}

	return spaces
}
