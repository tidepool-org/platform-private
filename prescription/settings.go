package prescription

import (
	"github.com/tidepool-org/platform/data/blood/glucose"
	"github.com/tidepool-org/platform/data/types/settings/pump"
	"github.com/tidepool-org/platform/structure"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
)

type InitialSettings struct {
	BloodGlucoseUnits            string                             `json:"bloodGlucoseUnits,omitempty" bson:"bloodGlucoseUnits,omitempty"`
	BasalRateSchedule            *pump.BasalRateStartArray          `json:"basalRateSchedule,omitempty" bson:"basalRateSchedule,omitempty"`
	BloodGlucoseSuspendThreshold *float64                           `json:"bloodGlucoseSuspendThreshold,omitempty" bson:"bloodGlucoseSuspendThreshold,omitempty"`
	BloodGlucoseTargetSchedule   *pump.BloodGlucoseTargetStartArray `json:"bloodGlucoseTargetSchedule,omitempty" bson:"bloodGlucoseTargetSchedule,omitempty"`
	CarbohydrateRatioSchedule    *pump.CarbohydrateRatioStartArray  `json:"carbohydrateRatioSchedule,omitempty" bson:"carbohydrateRatioSchedule,omitempty"`
	InsulinModel                 *string                            `json:"insulinModel,omitempty" bson:"insulinModel,omitempty"`
	InsulinSensitivitySchedule   *pump.InsulinSensitivityStartArray `json:"insulinSensitivitySchedule,omitempty" bson:"insulinSensitivitySchedule,omitempty"`
	BasalRateMaximum             *pump.BasalRateMaximum             `json:"basalRateMaximum,omitempty" bson:"basalRateMaximum,omitempty"`
	BolusAmountMaximum           *pump.BolusAmountMaximum           `json:"bolusAmountMaximum,omitempty" bson:"bolusAmountMaximum,omitempty"`
	PumpID                       string                             `json:"pumpId,omitempty" bson:"pumpId,omitempty"`
	CgmID                        string                             `json:"cgmId,omitempty" bson:"cgmId,omitempty"`
}

func AllowedInsulinModels() []string {
	return []string{
		pump.InsulinModelModelTypeRapidAdult,
		pump.InsulinModelModelTypeRapidChild,
	}
}

func (i *InitialSettings) Validate(validator structure.Validator) {
	validator.String("bloodGlucoseUnits", &i.BloodGlucoseUnits).EqualTo(glucose.MgdL)
	if i.BasalRateSchedule != nil {
		i.BasalRateSchedule.Validate(validator.WithReference("basalSchedule"))
	}
	if i.BloodGlucoseSuspendThreshold != nil {
		pump.ValidateBloodGlucoseSuspendThreshold(i.BloodGlucoseSuspendThreshold, &i.BloodGlucoseUnits, "bloodGlucoseSuspendThreshold", validator)
	}
	if i.BloodGlucoseTargetSchedule != nil {
		i.BloodGlucoseTargetSchedule.Validate(validator.WithReference("bloodGlucoseTargetSchedule"), &i.BloodGlucoseUnits)
	}
	if i.CarbohydrateRatioSchedule != nil {
		i.CarbohydrateRatioSchedule.Validate(validator.WithReference("carbohydrateRatioSchedule"))
	}
	if i.InsulinModel != nil {
		validator.String("insulinModel", i.InsulinModel).OneOf(AllowedInsulinModels()...)
	}
	if i.InsulinSensitivitySchedule != nil {
		i.InsulinSensitivitySchedule.Validate(validator.WithReference("insulinSensitivitySchedule"), &i.BloodGlucoseUnits)
	}
	if i.BasalRateMaximum != nil {
		i.BasalRateMaximum.Validate(validator.WithReference("basalRateMaximum"))
	}
	if i.BolusAmountMaximum != nil {
		i.BolusAmountMaximum.Validate(validator.WithReference("bolusAmountMaximum"))
	}
}

func (i *InitialSettings) ValidateAllRequired(validator structure.Validator) {
	if i.BasalRateSchedule == nil {
		validator.WithReference("basalSchedule").ReportError(structureValidator.ErrorValueEmpty())
	}
	if i.BloodGlucoseSuspendThreshold == nil {
		validator.WithReference("bloodGlucoseSuspendThreshold").ReportError(structureValidator.ErrorValueEmpty())
	}
	if i.BloodGlucoseTargetSchedule == nil {
		validator.WithReference("bloodGlucoseTargetSchedule").ReportError(structureValidator.ErrorValueEmpty())
	}
	if i.CarbohydrateRatioSchedule == nil {
		validator.WithReference("carbohydrateRatioSchedule").ReportError(structureValidator.ErrorValueEmpty())
	}
	if i.InsulinModel == nil {
		validator.WithReference("insulinModel").ReportError(structureValidator.ErrorValueEmpty())
	}
	if i.InsulinSensitivitySchedule == nil {
		validator.WithReference("insulinSensitivitySchedule").ReportError(structureValidator.ErrorValueEmpty())
	}
	if i.BasalRateMaximum == nil {
		validator.WithReference("basalRateMaximum").ReportError(structureValidator.ErrorValueEmpty())
	}
	if i.BolusAmountMaximum == nil {
		validator.WithReference("bolusAmountMaximum").ReportError(structureValidator.ErrorValueEmpty())
	}
	if i.PumpID == "" {
		validator.WithReference("pumpId").ReportError(structureValidator.ErrorValueEmpty())
	}
	if i.CgmID == "" {
		validator.WithReference("cgmId").ReportError(structureValidator.ErrorValueEmpty())
	}
}