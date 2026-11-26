import React, { useCallback, useEffect, useState } from 'react';
import { css } from '@emotion/css';
import {
  Alert,
  Button,
  Card,
  Grid,
  LoadingPlaceholder,
  Tab,
  TabsBar,
  useStyles2,
} from '@grafana/ui';
import { lastValueFrom, Observable } from 'rxjs';
import { FetchResponse, getBackendSrv } from '@grafana/runtime';
import { GrafanaTheme2, TimeRange } from '@grafana/data';
import { ArrowLeft, ExternalLink, RefreshCw } from 'lucide-react';
import { ExecutionPlanViewer } from '../ExecutionPlanTree/plan';
import { BlockingChainComponent } from './blockingChain';
import { QueryDetails } from './types';
import { TempoTraceLink } from '../TempoTraceLink';
import { QueryHistoryChart } from '../QueryMetrics/QueryHistoryChart';
import { SqlQueryCard } from '../SqlQueryCard';

export type DetailsTabId = 'overview' | 'plan' | 'blocking' | 'history';

const getDrawerStyles = (theme: GrafanaTheme2) => ({
  headerBar: css`
    display: flex;
    align-items: center;
    justify-content: space-between;
    flex-wrap: wrap;
    gap: ${theme.spacing(1)};
    margin-bottom: ${theme.spacing(2)};
  `,
  headerActions: css`
    display: flex;
    align-items: center;
    gap: ${theme.spacing(1)};
    flex-wrap: wrap;
  `,
  tabContent: css`
    margin-top: ${theme.spacing(2)};
  `,
  metaLabel: css`
    color: ${theme.colors.text.secondary};
    font-size: ${theme.typography.bodySmall.fontSize};
    margin: 0 0 2px 0;
  `,
  metaValue: css`
    color: ${theme.colors.text.primary};
    font-family: ${theme.typography.fontFamilyMonospace};
    font-size: ${theme.typography.bodySmall.fontSize};
    word-break: break-all;
    margin: 0 0 ${theme.spacing(1)} 0;
  `,
  sectionGap: css`
    margin-bottom: ${theme.spacing(2)};
  `,
  kpiRow: css`
    display: flex;
    flex-wrap: wrap;
    gap: ${theme.spacing(1.5)};
    margin-bottom: ${theme.spacing(2)};
  `,
  kpiPill: css`
    padding: ${theme.spacing(1, 1.5)};
    background: ${theme.colors.background.secondary};
    border: 1px solid ${theme.colors.border.weak};
    border-radius: ${theme.shape.radius.default};
    min-width: 110px;
  `,
});

export const QueryDetailsComponent: React.FC<{
  sampleID: string;
  snapID: string;
  timeRange: TimeRange;
  server: string;
  initialTab?: DetailsTabId;
  onSelectSample?: (snapId: string, sampleId: string) => void;
  onOpenSnapshot?: (snapId: string) => void;
  onOpenQueryMetrics?: (queryHash: string) => void;
}> = ({
  sampleID,
  snapID,
  timeRange,
  server,
  initialTab = 'overview',
  onSelectSample,
  onOpenSnapshot,
  onOpenQueryMetrics,
}) => {
  const styles = useStyles2(getDrawerStyles);

  const [historyStack, setHistoryStack] = useState<Array<{ snapId: string; sampleId: string }>>([]);
  const [activeTarget, setActiveTarget] = useState<{ snapId: string; sampleId: string }>({
    snapId: snapID,
    sampleId: sampleID,
  });
  const [activeTab, setActiveTab] = useState<DetailsTabId>(initialTab);
  const [queryDetails, setQueryDetails] = useState<QueryDetails | null>(null);
  const [detailsLoading, setDetailsLoading] = useState(false);
  const [detailsError, setDetailsError] = useState<string | null>(null);

  // Sync when external selection changes
  useEffect(() => {
    setActiveTarget({ snapId: snapID, sampleId: sampleID });
    setHistoryStack([]);
    setActiveTab(initialTab);
  }, [snapID, sampleID, initialTab]);

  const fetchDetails = useCallback(async () => {
    if (!activeTarget.sampleId || !activeTarget.snapId) {
      return;
    }
    try {
      setDetailsLoading(true);
      setDetailsError(null);
      const params = new URLSearchParams({
        sampleId: activeTarget.sampleId,
        snapId: activeTarget.snapId,
      });
      const response: Observable<FetchResponse<QueryDetails>> = getBackendSrv().fetch({
        url: `/api/plugins/guilhermearpassos-sqlsights-app/resources/getQueryDetails?${params.toString()}`,
      });
      const result = await lastValueFrom(response);
      setQueryDetails(result.data);
    } catch (err) {
      setDetailsError(err instanceof Error ? err.message : 'Failed to load query details');
      setQueryDetails(null);
    } finally {
      setDetailsLoading(false);
    }
  }, [activeTarget.sampleId, activeTarget.snapId]);

  useEffect(() => {
    fetchDetails();
  }, [fetchDetails]);

  const handleNavigateInChain = (nextSnapId: string, nextSampleId: string) => {
    setHistoryStack((prev) => [...prev, activeTarget]);
    setActiveTarget({ snapId: nextSnapId, sampleId: nextSampleId });
    onSelectSample?.(nextSnapId, nextSampleId);
  };

  const handleBack = () => {
    if (historyStack.length === 0) {
      return;
    }
    const prevTarget = historyStack[historyStack.length - 1];
    setHistoryStack((prev) => prev.slice(0, -1));
    setActiveTarget(prevTarget);
    onSelectSample?.(prevTarget.snapId, prevTarget.sampleId);
  };

  const warningsCount = queryDetails?.plan?.warnings?.length ?? 0;
  const blockingRootsCount = queryDetails?.blocking_chain?.roots?.length ?? 0;
  const sample = queryDetails?.query_sample;

  return (
    <div>
      <div className={styles.headerBar}>
        <div className={styles.headerActions}>
          {historyStack.length > 0 && (
            <Button size="sm" variant="secondary" onClick={handleBack}>
              <ArrowLeft size={14} />
              <span style={{ marginLeft: 4 }}>Back</span>
            </Button>
          )}
          {sample?.trace_id && <TempoTraceLink traceId={sample.trace_id} timeRange={timeRange} />}
        </div>
        <div className={styles.headerActions}>
          {onOpenSnapshot && activeTarget.snapId && (
            <Button size="sm" variant="secondary" fill="outline" onClick={() => onOpenSnapshot(activeTarget.snapId)}>
              <ExternalLink size={13} />
              <span style={{ marginLeft: 4 }}>Open Snapshot</span>
            </Button>
          )}
          <Button size="sm" variant="secondary" fill="outline" onClick={fetchDetails} disabled={detailsLoading}>
            <RefreshCw size={13} />
            <span style={{ marginLeft: 4 }}>Refresh</span>
          </Button>
        </div>
      </div>

      {detailsError && (
        <Alert title="Error loading sample details" severity="error" className={styles.sectionGap}>
          <div>{detailsError}</div>
          <Button size="sm" variant="secondary" onClick={fetchDetails} style={{ marginTop: 8 }}>
            Retry
          </Button>
        </Alert>
      )}

      {detailsLoading && !queryDetails && <LoadingPlaceholder text="Loading sample details..." />}

      {queryDetails && sample && (
        <>
          <TabsBar>
            <Tab label="Overview" active={activeTab === 'overview'} onChangeTab={() => setActiveTab('overview')} />
            <Tab
              label="Execution Plan"
              counter={warningsCount > 0 ? warningsCount : undefined}
              active={activeTab === 'plan'}
              onChangeTab={() => setActiveTab('plan')}
            />
            <Tab
              label="Blocking Chain"
              counter={blockingRootsCount > 0 ? blockingRootsCount : undefined}
              active={activeTab === 'blocking'}
              onChangeTab={() => setActiveTab('blocking')}
            />
            <Tab label="History" active={activeTab === 'history'} onChangeTab={() => setActiveTab('history')} />
          </TabsBar>

          <div className={styles.tabContent}>
            {activeTab === 'overview' && (
              <div>
                <div className={styles.kpiRow}>
                  <div className={styles.kpiPill}>
                    <p className={styles.metaLabel}>Database</p>
                    <p className={styles.metaValue} style={{ marginBottom: 0 }}>
                      {sample.database || '-'}
                    </p>
                  </div>
                  <div className={styles.kpiPill}>
                    <p className={styles.metaLabel}>Status</p>
                    <p className={styles.metaValue} style={{ marginBottom: 0 }}>
                      {sample.status || '-'}
                    </p>
                  </div>
                  <div className={styles.kpiPill}>
                    <p className={styles.metaLabel}>Elapsed</p>
                    <p className={styles.metaValue} style={{ marginBottom: 0 }}>
                      {sample.execution_time || '-'}
                    </p>
                  </div>
                  <div className={styles.kpiPill}>
                    <p className={styles.metaLabel}>Wait Event</p>
                    <p className={styles.metaValue} style={{ marginBottom: 0 }}>
                      {sample.wait_event || 'CPU / Running'}
                    </p>
                  </div>
                  <div className={styles.kpiPill}>
                    <p className={styles.metaLabel}>Session ID</p>
                    <p className={styles.metaValue} style={{ marginBottom: 0 }}>
                      {sample.sid}
                    </p>
                  </div>
                </div>

                <SqlQueryCard sql={sample.query} heading="SQL Query" className={styles.sectionGap} />

                <Card className={styles.sectionGap}>
                  <Card.Heading>Session & Client Metadata</Card.Heading>
                  <Card.Description>
                    <Grid columns={3}>
                      <div>
                        <p className={styles.metaLabel}>Login</p>
                        <p className={styles.metaValue}>{sample.user || '-'}</p>
                      </div>
                      <div>
                        <p className={styles.metaLabel}>Program Name</p>
                        <p className={styles.metaValue}>{sample.session_program_name || '-'}</p>
                      </div>
                      <div>
                        <p className={styles.metaLabel}>Session Status</p>
                        <p className={styles.metaValue}>{sample.session_status || '-'}</p>
                      </div>
                      <div>
                        <p className={styles.metaLabel}>Client Host</p>
                        <p className={styles.metaValue}>{sample.session_host || '-'}</p>
                      </div>
                      <div>
                        <p className={styles.metaLabel}>Client IP</p>
                        <p className={styles.metaValue}>{sample.session_client_ip || '-'}</p>
                      </div>
                      <div>
                        <p className={styles.metaLabel}>Login Time</p>
                        <p className={styles.metaValue}>{sample.session_login_time || '-'}</p>
                      </div>
                    </Grid>
                  </Card.Description>
                </Card>

                <Card>
                  <Card.Heading>Identifiers</Card.Heading>
                  <Card.Description>
                    <Grid columns={2}>
                      <div>
                        <p className={styles.metaLabel}>Snapshot ID</p>
                        <p className={styles.metaValue}>{sample.snap_id}</p>
                      </div>
                      <div>
                        <p className={styles.metaLabel}>Sample ID</p>
                        <p className={styles.metaValue}>{sample.sample_id}</p>
                      </div>
                      <div>
                        <p className={styles.metaLabel}>Query Hash</p>
                        <p className={styles.metaValue}>{sample.query_hash || '-'}</p>
                      </div>
                      <div>
                        <p className={styles.metaLabel}>SQL Handle</p>
                        <p className={styles.metaValue}>{sample.sql_handle || '-'}</p>
                      </div>
                    </Grid>
                  </Card.Description>
                </Card>
              </div>
            )}

            {activeTab === 'plan' && (
              <div>
                {queryDetails.plan ? (
                  <ExecutionPlanViewer executionPlan={queryDetails.plan} />
                ) : (
                  <Alert title="No Execution Plan" severity="info">
                    No execution plan was found for this sample.
                  </Alert>
                )}
              </div>
            )}

            {activeTab === 'blocking' && (
              <div>
                {queryDetails.blocking_chain && blockingRootsCount > 0 ? (
                  <BlockingChainComponent
                    chain={queryDetails.blocking_chain}
                    currentSampleId={activeTarget.sampleId}
                    timeRange={timeRange}
                    onSelectSample={handleNavigateInChain}
                  />
                ) : (
                  <Alert title="No Blocking Chain" severity="info">
                    This sample is neither blocking other sessions nor blocked by another session.
                  </Alert>
                )}
              </div>
            )}

            {activeTab === 'history' && (
              <Card>
                <Card.Heading>Executions &amp; Performance History</Card.Heading>
                <Card.Description>
                  {sample.query_hash ? (
                    <QueryHistoryChart
                      server={server}
                      queryHash={sample.query_hash}
                      timeRange={timeRange}
                      initialMetric="execution_count"
                      height={260}
                      onOpenQueryMetrics={onOpenQueryMetrics}
                    />
                  ) : (
                    <Alert title="No Query Hash" severity="info">
                      This sample does not have a SQL Server query hash (for example, a sleeping session or utility
                      command), so execution counter history is unavailable.
                    </Alert>
                  )}
                </Card.Description>
              </Card>
            )}
          </div>
        </>
      )}
    </div>
  );
};
