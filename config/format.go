package config

import (
	"regexp"

	"github.com/xeipuuv/gojsonschema"
)

// https://aws.amazon.com/rds/instance-types/
type RDSInstanceTypeFormat struct{}

var rdsInstanceTypeFormatRegex = regexp.MustCompile(`^db\.[\w\d]+\.[\w\d]+$`)

func (f RDSInstanceTypeFormat) IsFormat(input interface{}) bool {
	v, ok := input.(string)
	if !ok {
		return false
	}
	return rdsInstanceTypeFormatRegex.MatchString(v)
}

func init() {
	gojsonschema.FormatCheckers.Add("rds-instance-type", RDSInstanceTypeFormat{})
}
