package query

import (
	"context"
	"errors"
	"fmt"
	"github.com/guilhermearpassos/database-monitoring/internal/common/custom_errors"
	"github.com/guilhermearpassos/database-monitoring/internal/services/collector/adapters/parsers"
	"github.com/guilhermearpassos/database-monitoring/internal/services/collector/domain"
	"github.com/guilhermearpassos/database-monitoring/internal/services/common_domain"
	"github.com/guilhermearpassos/database-monitoring/internal/services/common_domain/converters"
	dbmv1 "github.com/guilhermearpassos/database-monitoring/proto/database_monitoring/v1"
)

type GetQuerySampleDetailsHandler struct {
	repo domain.SampleRepository
}

func NewGetQuerySampleDetailsHandler(repo domain.SampleRepository) GetQuerySampleDetailsHandler {
	return GetQuerySampleDetailsHandler{repo: repo}
}
func (h *GetQuerySampleDetailsHandler) Handle(ctx context.Context, snapID string, sampleID []byte) (*dbmv1.GetSampleDetailsResponse, error) {
	snap, err := h.repo.GetSnapshot(ctx, snapID)
	if err != nil {
		return nil, fmt.Errorf("get snapshot %s error: %w", snapID, err)
	}
	sampleMap := make(map[string][]*common_domain.QuerySample, len(snap.Samples))
	var baseQuery *common_domain.QuerySample
	found := false
	for _, sample := range snap.Samples {
		if string(sample.Id) == string(sampleID) {
			found = true
			baseQuery = sample
		}
		if _, ok := sampleMap[sample.Session.SessionID]; ok {
			sampleMap[sample.Session.SessionID] = append(sampleMap[sample.Session.SessionID], sample)
		} else {
			sampleMap[sample.Session.SessionID] = []*common_domain.QuerySample{sample}
		}
	}
	if !found || baseQuery == nil {
		return nil, custom_errors.NotFoundErr{Message: fmt.Sprintf("sample %s not found", snapID)}
	}
	roots := make([]string, 0)
	if baseQuery.IsBlocked {
		traveled := make(map[string]struct{}, 0)
		roots2 := h.searchForRoot(sampleMap, baseQuery, traveled)
		roots = append(roots, roots2...)
	}
	if baseQuery.IsBlocker && !baseQuery.IsBlocked {
		roots = append(roots, baseQuery.Session.SessionID)
	}
	participants := make([]*dbmv1.BlockChain_BlockingNode, len(roots))
	traveled := make(map[string]struct{}, 0)
	for i, root := range roots {
		participants[i] = h.buildblockNode(sampleMap, root, traveled)
	}
	planFound := true
	plan, err := h.repo.GetExecutionPlan(ctx, baseQuery.PlanHandle, &common_domain.ServerMeta{
		Host: snap.SnapInfo.Server.Host,
		Type: "mssql",
	})
	if err != nil {
		if !errors.As(err, &custom_errors.NotFoundErr{}) {
			return nil, fmt.Errorf("get snapshot %s error: %w", snapID, err)
		}
		planFound = false
	}
	var protoParsedPlan *dbmv1.ParsedExecutionPlan
	if planFound {
		parsedPlan, err2 := parsers.ParseExecutionPlan(plan.XmlData)
		if err2 != nil {
			return nil, fmt.Errorf("parse execution plan error: %w", err2)
		}
		protoParsedPlan, err = parsers.PlanToProto(plan.PlanHandle, common_domain.ServerMeta{
			Host: snap.SnapInfo.Server.Host,
			Type: snap.SnapInfo.Server.Type,
		}, parsedPlan)
		if err != nil {
			return nil, fmt.Errorf("parsed execution to proto: %w", err)
		}
	}

	return &dbmv1.GetSampleDetailsResponse{
		QuerySample: converters.SampleToProto(baseQuery),
		ParsedPlan:  protoParsedPlan,
		BlockChain:  &dbmv1.BlockChain{Roots: participants},
	}, nil

}

func (h *GetQuerySampleDetailsHandler) searchForRoot(sampleMap map[string][]*common_domain.QuerySample, currentQuery *common_domain.QuerySample, traveled map[string]struct{}) []string {

	traveled[currentQuery.Session.SessionID] = struct{}{}
	roots2 := make([]string, 0)
	if !currentQuery.IsBlocked {

		roots2 = append(roots2, currentQuery.Session.SessionID)

		return roots2
	}
	blockingSessionSamples, ok := sampleMap[currentQuery.Block.BlockedBy]
	if !ok {
		roots2 = append(roots2, currentQuery.Session.SessionID)
	}
	for _, blockingSessionSample := range blockingSessionSamples {
		if _, tok := traveled[blockingSessionSample.Session.SessionID]; tok {
			roots2 = append(roots2, blockingSessionSample.Session.SessionID)
			continue
		}
		root := h.searchForRoot(sampleMap, blockingSessionSample, traveled)
		roots2 = append(roots2, root...)
	}
	return roots2
}

func (h *GetQuerySampleDetailsHandler) buildblockNode(sampleMap map[string][]*common_domain.QuerySample, sessionID string, traveled map[string]struct{}) *dbmv1.BlockChain_BlockingNode {
	if _, ok := traveled[sessionID]; ok {
		return nil
	}
	traveled[sessionID] = struct{}{}
	samplesForSession, ok := sampleMap[sessionID]
	if !ok {
		return nil
	}
	sample := samplesForSession[0]
	childNodes := make([]*dbmv1.BlockChain_BlockingNode, 0, len(sample.Block.BlockedSessions))
	for _, s := range sample.Block.BlockedSessions {
		node := h.buildblockNode(sampleMap, s, traveled)
		if node == nil {
			continue
		}
		childNodes = append(childNodes, node)
	}
	return &dbmv1.BlockChain_BlockingNode{
		QuerySample: converters.SampleToProto(sample),
		ChildNodes:  childNodes,
	}
}
