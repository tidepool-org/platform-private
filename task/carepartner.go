package task

import (
	"fmt"
	"time"

	"github.com/tidepool-org/platform/pointer"
)

const CarePartnerType = "org.tidepool.carepartner"

func NewCarePartnerTaskCreate() *TaskCreate {
	fmt.Println("!!! NewCarePartnerTaskCreate !!!")
	return &TaskCreate{
		Name:          pointer.FromAny(CarePartnerType),
		Type:          CarePartnerType,
		AvailableTime: &time.Time{},
		Data:          map[string]interface{}{},
	}
}
