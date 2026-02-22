package app

import (
	"etruscan/internal/usecases"
)

type Dependencies struct {
	AuthUseCase          *usecases.AuthUseCase
	UserUseCase          *usecases.UserUseCase
	ApproverGroupUseCase *usecases.ApproverGroupUseCase
	FlagUseCase          *usecases.FlagUseCase
	ExperimentUseCase    *usecases.ExperimentUseCase
	DecideUseCase        *usecases.DecideUseCase
	EventUseCase         *usecases.EventUseCase
	EventTypeUseCase     *usecases.EventTypeUseCase
	MetricUseCase        *usecases.MetricUseCase
	ReportUseCase        *usecases.ReportUseCase
}
