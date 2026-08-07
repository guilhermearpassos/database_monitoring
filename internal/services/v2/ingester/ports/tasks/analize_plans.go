package tasks

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/guilhermearpassos/database-monitoring/internal/runtimes"
	"github.com/guilhermearpassos/database-monitoring/internal/services/v2/ingester/app"
	"github.com/guilhermearpassos/database-monitoring/internal/services/v2/ingester/domain"
	"github.com/guilhermearpassos/database-monitoring/internal/services/v2/ingester/ports/tasks/parsers"
)

type AnalizePlansTask struct {
	interval time.Duration
	logger   *slog.Logger
	app      app.Application
}

func NewAnalizePlansTask(application app.Application, interval time.Duration, logger *slog.Logger) *AnalizePlansTask {
	return &AnalizePlansTask{
		interval: interval,
		logger:   logger,
		app:      application,
	}
}

var _ runtimes.Task = (*AnalizePlansTask)(nil)

func (c AnalizePlansTask) Name() string {
	return "AnalizePlansTask"
}

func (c AnalizePlansTask) Interval() time.Duration {
	return c.interval
}

func (c AnalizePlansTask) Run(ctx context.Context) error {
	plansToAnalize, err := c.app.Query.GetUnanalizedPlans.Handle(ctx, 50)
	if err != nil {
		return fmt.Errorf("get unanalized plans: %w", err)
	}
	for _, plan := range plansToAnalize {
		parsedPlan, err2 := parsers.ParseExecutionPlan(plan.XmlData)
		if err2 != nil {
			err = errors.Join(err, err2)
		}
		MissingIndexes := 0
		ImplicitConversions := 0
		LargeTableScans := 0
		for _, b := range parsedPlan.Batches {
			for _, s := range b.Statements.StmtSimple {
				if s.QueryPlan == nil {
					continue
				}
				if x := s.QueryPlan.MissingIndexes; x != nil {

					MissingIndexes += len(x.MissingIndexGroups)
				}
				ImplicitConversions += len(s.QueryPlan.PlanAffectingConvert)
			}
		}
		res := domain.PlanAnalisysResults{
			MissingIndexes:      MissingIndexes,
			ImplicitConversions: ImplicitConversions,
			LargeTableScans:     LargeTableScans,
		}
		err2 = c.app.Command.SetPlanAnalisys.Handle(ctx, plan.PlanHandle, res)
		if err2 != nil {
			err = errors.Join(err, err2)
		}
	}
	return err
}

func (c AnalizePlansTask) Opts() runtimes.TaskOpts {
	return runtimes.TaskOpts{Overlap: runtimes.OverlapOptDrop}
}
