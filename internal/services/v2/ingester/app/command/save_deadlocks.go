package command

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/guilhermearpassos/database-monitoring/internal/deadlock"
	"github.com/guilhermearpassos/database-monitoring/internal/services/common_domain"
	"github.com/guilhermearpassos/database-monitoring/internal/services/v2/ingester/domain"
)

type SaveDeadlocksHandler struct {
	repo domain.DeadlockRepository
}

func NewSaveDeadlocksHandler(repo domain.DeadlockRepository) *SaveDeadlocksHandler {
	return &SaveDeadlocksHandler{repo: repo}
}

func (h *SaveDeadlocksHandler) Handle(ctx context.Context, server common_domain.ServerMeta, deadlocks []*common_domain.DeadlockReport) (int, error) {
	if h.repo == nil {
		return 0, fmt.Errorf("deadlock repository not configured")
	}
	for _, d := range deadlocks {
		if d == nil || d.XMLReport == "" {
			continue
		}
		if d.Fingerprint == "" || d.Pattern == "" || d.XMLHash == "" {
			parsed, err := deadlock.ParseXML(d.XMLReport)
			if err != nil {
				slog.WarnContext(ctx, "failed to parse deadlock XML in ingester", "server", server.Host, "deadlock_id", d.ID, "error", err)
			} else if parsed != nil {
				d.XMLHash = parsed.XMLHash
				d.Fingerprint = parsed.Fingerprint
				d.Pattern = parsed.Diagnosis.Pattern
				d.Databases = parsed.Databases
				d.VictimSPID = parsed.VictimSPID
				d.MaxWaitMs = parsed.MaxWaitMs
				d.ProcessCount = int32(len(parsed.Processes))
				d.ResourceCount = int32(len(parsed.Resources))
			}
		}
	}
	return h.repo.SaveDeadlocks(ctx, server, deadlocks)
}
