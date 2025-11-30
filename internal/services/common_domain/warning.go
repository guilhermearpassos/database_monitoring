package common_domain

import (
	"fmt"

	"github.com/google/uuid"
	dbmv1 "github.com/guilhermearpassos/database-monitoring/proto/database_monitoring/v1"
)

type Warning struct {
	Id          string
	WarningType string
	WarningData *dbmv1.Warning
}

func NewWarning(warningData *dbmv1.Warning) *Warning {
	if warningData.Id == "" {
		warningData.Id = uuid.NewString()
	}
	return &Warning{
		Id:          warningData.Id,
		WarningType: fmt.Sprintf("%T", warningData.Type),
		WarningData: warningData,
	}
}
