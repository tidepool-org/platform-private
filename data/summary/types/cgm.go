package types

import (
	"github.com/tidepool-org/platform/data/types/blood/glucose/continuous"
)

type GlucosePeriods map[string]*GlucosePeriod

type GlucoseStats struct {
	Periods       GlucosePeriods `json:"periods,omitempty" bson:"periods,omitempty"`
	OffsetPeriods GlucosePeriods `json:"offsetPeriods,omitempty" bson:"offsetPeriods,omitempty"`
	TotalHours    int            `json:"totalHours,omitempty" bson:"totalHours,omitempty"`
}

type CGMStats struct {
	GlucoseStats
}

func (*CGMStats) GetType() string {
	return SummaryTypeCGM
}

func (*CGMStats) GetDeviceDataTypes() []string {
	return []string{continuous.Type}
}
