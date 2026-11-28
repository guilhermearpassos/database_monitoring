package converters

import (
	"github.com/guilhermearpassos/database-monitoring/internal/services/common_domain"
	dbmv1 "github.com/guilhermearpassos/database-monitoring/proto/database_monitoring/v1"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func DeadlockReportToProto(d *common_domain.DeadlockReport) *dbmv1.DeadlockReport {
	if d == nil {
		return nil
	}
	var ts *timestamppb.Timestamp
	if !d.OccurredAt.IsZero() {
		ts = timestamppb.New(d.OccurredAt)
	}
	return &dbmv1.DeadlockReport{
		Id:            d.ID,
		Server:        &dbmv1.ServerMetadata{Host: d.Server.Host, Type: d.Server.Type},
		OccurredAt:    ts,
		XmlReport:     d.XMLReport,
		XmlHash:       d.XMLHash,
		Fingerprint:   d.Fingerprint,
		Pattern:       d.Pattern,
		Databases:     d.Databases,
		VictimSpid:    d.VictimSPID,
		MaxWaitMs:     d.MaxWaitMs,
		ProcessCount:  d.ProcessCount,
		ResourceCount: d.ResourceCount,
	}
}

func DeadlockReportToDomain(p *dbmv1.DeadlockReport) *common_domain.DeadlockReport {
	if p == nil {
		return nil
	}
	var srv common_domain.ServerMeta
	if p.GetServer() != nil {
		srv = common_domain.ServerMeta{
			Host: p.GetServer().GetHost(),
			Type: p.GetServer().GetType(),
		}
	}
	return &common_domain.DeadlockReport{
		ID:            p.GetId(),
		Server:        srv,
		OccurredAt:    p.GetOccurredAt().AsTime(),
		XMLReport:     p.GetXmlReport(),
		XMLHash:       p.GetXmlHash(),
		Fingerprint:   p.GetFingerprint(),
		Pattern:       p.GetPattern(),
		Databases:     p.GetDatabases(),
		VictimSPID:    p.GetVictimSpid(),
		MaxWaitMs:     p.GetMaxWaitMs(),
		ProcessCount:  p.GetProcessCount(),
		ResourceCount: p.GetResourceCount(),
	}
}
