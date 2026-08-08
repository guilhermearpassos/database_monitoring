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
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

type AnalizePlansTask struct {
	interval  time.Duration
	logger    *slog.Logger
	app       app.Application
	batchSize int
}

func NewAnalizePlansTask(application app.Application, interval time.Duration, batchSize int, logger *slog.Logger) *AnalizePlansTask {
	return &AnalizePlansTask{
		interval:  interval,
		logger:    logger,
		app:       application,
		batchSize: batchSize,
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
	span := trace.SpanFromContext(ctx)

	plansToAnalize, err := c.app.Query.GetUnanalizedPlans.Handle(ctx, c.batchSize)
	if err != nil {
		return fmt.Errorf("get unanalized plans: %w", err)
	}
	span.SetAttributes(attribute.Int("plan_count", len(plansToAnalize)))
	var results []domain.PlanAnalisysBatchResult
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
		results = append(results, domain.PlanAnalisysBatchResult{
			PlanHandle: plan.PlanHandle,
			Results: domain.PlanAnalisysResults{
				MissingIndexes:      MissingIndexes,
				ImplicitConversions: ImplicitConversions,
				LargeTableScans:     LargeTableScans,
			},
		})
	}
	if len(results) > 0 {
		err2 := c.app.Command.SetPlanAnalisys.Handle(ctx, results)
		if err2 != nil {
			err = errors.Join(err, err2)
		}
	}
	return err
}

func (c AnalizePlansTask) Opts() runtimes.TaskOpts {
	return runtimes.TaskOpts{Overlap: runtimes.OverlapOptDrop}
}
