package test

import (
	"time"

	"github.com/tidepool-org/platform/dexcom"
)

func RandomDataRangeResponse() *dexcom.DataRangeResponse {
	datum := dexcom.NewDataRangeResponse()
	datum.Calibrations = RandomDataRange()
	datum.Egvs = RandomDataRange()
	datum.Events = RandomDataRange()
	return datum
}

func RandomDataRangeResponseWithDate(seed time.Time) *dexcom.DataRangeResponse {
	datum := dexcom.NewDataRangeResponse()
	datum.Calibrations = dataRange(seed)
	datum.Egvs = dataRange(seed)
	datum.Events = dataRange(seed)
	return datum
}

func CloneDataRangeResponse(datum *dexcom.DataRangeResponse) *dexcom.DataRangeResponse {
	if datum == nil {
		return nil
	}
	clone := dexcom.NewDataRangeResponse()
	clone.Calibrations = CloneDataRange(datum.Calibrations)
	clone.Events = CloneDataRange(datum.Events)
	clone.Egvs = CloneDataRange(datum.Egvs)
	return clone
}

func RandomDataRange() *dexcom.DataRange {
	datum := dexcom.NewDataRange()
	datum.End = RandomDateRange()
	datum.Start = RandomDateRange()
	return datum
}

func dataRange(seed time.Time) *dexcom.DataRange {
	datum := dexcom.NewDataRange()
	datum.End.DisplayTime.Time = seed
	datum.End.SystemTime.Time = seed
	datum.Start.DisplayTime.Time = seed.Add(-12 * time.Hour)
	datum.Start.SystemTime.Time = seed.Add(-12 * time.Hour)
	return datum
}

func CloneDataRange(datum *dexcom.DataRange) *dexcom.DataRange {
	if datum == nil {
		return nil
	}
	clone := dexcom.NewDataRange()
	clone.End = CloneDateRange(datum.End)
	clone.Start = CloneDateRange(datum.Start)
	return clone
}

func RandomDateRange() *dexcom.DateRange {
	datum := dexcom.NewDateRange()
	datum.DisplayTime = RandomDisplayTime()
	datum.SystemTime = RandomSystemTime()
	return datum
}

func CloneDateRange(datum *dexcom.DateRange) *dexcom.DateRange {
	if datum == nil {
		return nil
	}
	clone := dexcom.NewDateRange()
	clone.DisplayTime = CloneTime(datum.DisplayTime)
	clone.SystemTime = CloneTime(datum.SystemTime)
	return clone
}