import React, { useCallback, useEffect, useMemo, useState } from 'react';
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
import { FetchResponse, getBackendSrv, getDataSourceSrv } from '@grafana/runtime';
import {
  CoreApp,
  DataFrame,
  dataFrameFromJSON,
  DataQueryRequest,
  DataSourceApi,
  EventBusSrv,
  GrafanaTheme2,
  LoadingState,
  TimeRange,
} from '@grafana/data';
import { DataQueryResponse } from '@grafana/data/dist/types/types/datasource';
import { ArrowLeft, Check, Copy, ExternalLink, RefreshCw } from 'lucide-react';
import { ExecutionPlanViewer } from '../ExecutionPlanTree/plan';
import { BlockingChainComponent } from './blockingChain';
import { QueryDetails } from './types';
import { MyQuery } from '../../nested-datasource/types';
import { MyGraph } from '../../pages/graph';
import { formatSQL } from '../../utils/formatters';
import { copyToClipboard, formatSQLPlain } from '../../utils/sqlHelpers';
import { TempoTraceLink } from '../TempoTraceLink';

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
  sqlBox: css`
    position: relative;
    background: ${theme.colors.background.canvas};
    border: 1px solid ${theme.colors.border.weak};
    border-radius: ${theme.shape.radius.default};
    padding: ${theme.spacing(1.5)};
    font-family: ${theme.typography.fontFamilyMonospace};
    font-size: ${theme.typography.bodySmall.fontSize};
    white-space: pre-wrap;
    word-break: break-word;
    max-height: 320px;
    overflow: auto;
    line-height: 1.5;
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
}> = ({ sampleID, snapID, timeRange, server, initialTab = 'overview', onSelectSample, onOpenSnapshot }) => {
  const styles = useStyles2(getDrawerStyles);
  const panelEventBus = useMemo(() => new EventBusSrv(), []);

  const [historyStack, setHistoryStack] = useState<Array<{ snapId: string; sampleId: string }>>([]);
  const [activeTarget, setActiveTarget] = useState<{ snapId: string; sampleId: string }>({
    snapId: snapID,
    sampleId: sampleID,
  });
  const [activeTab, setActiveTab] = useState<DetailsTabId>(initialTab);
  const [queryDetails, setQueryDetails] = useState<QueryDetails | null>(null);
  const [detailsLoading, setDetailsLoading] = useState(false);
  const [detailsError, setDetailsError] = useState<string | null>(null);
  const [copiedSql, setCopiedSql] = useState(false);

  const [chartFrames, setChartFrames] = useState<DataFrame[]>([]);
  const [chartLoading, setChartLoading] = useState(false);
  const [datasource, setDatasource] = useState<DataSourceApi | null>(null);

  // Sync when external selection changes
  useEffect(() => {
    setActiveTarget({ snapId: snapID, sampleId: sampleID });
    setHistoryStack([]);
    setActiveTab(initialTab);
  }, [snapID, sampleID, initialTab]);

  useEffect(() => {
    const initDatasource = async () => {
      try {
        const dsSrv = getDataSourceSrv();
        if (!dsSrv) {
          return;
        }
        const ds = await dsSrv.get({ type: 'guilhermearpassos-sqlsights-datasource' });
        setDatasource(ds);
      } catch {
        // optional for history chart
      }
    };
    initDatasource();
  }, []);

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

  const loadHistoryChart = useCallback(
    async (qd: QueryDetails) => {
      if (!datasource || !qd.query_sample.query_hash || !server) {
        setChartFrames([]);
        return;
      }
      try {
        setChartLoading(true);
        const query: MyQuery = {
          refId: 'A',
          database: server,
          hide: false,
          datasource: { type: datasource.type, uid: datasource.uid },
          queryType: 'metrics_series',
          queryHash: qd.query_sample.query_hash,
          metrics: ['executionCount'],
        };
        const queryRequest: DataQueryRequest<MyQuery> = {
          app: CoreApp.Explore,
          requestId: `details_${Date.now()}`,
          timezone: 'browser',
          panelId: 1,
          dashboardUID: '',
          range: timeRange,
          timeInfo: '',
          interval: '1m',
          intervalMs: 60000,
          targets: [query],
          maxDataPoints: 1000,
          scopedVars: {},
          startTime: Date.now(),
          liveStreaming: false,
        };
        const resultObservable = datasource.query(queryRequest);
        const result: DataQueryResponse =
          resultObservable instanceof Observable ? await lastValueFrom(resultObservable) : await resultObservable;
        if (result.data && Array.isArray(result.data)) {
          const frames = result.data
            .map((rawFrame: any) => {
              try {
                return dataFrameFromJSON({ schema: rawFrame.schema, data: rawFrame.data });
              } catch {
                return null;
              }
            })
            .filter((f): f is DataFrame => f !== null);
          setChartFrames(frames);
        } else {
          setChartFrames([]);
        }
      } catch {
        setChartFrames([]);
      } finally {
        setChartLoading(false);
      }
    },
    [datasource, server, timeRange]
  );

  useEffect(() => {
    if (queryDetails && activeTab === 'history') {
      loadHistoryChart(queryDetails);
    }
  }, [queryDetails, activeTab, loadHistoryChart]);

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

  const handleCopySql = async () => {
    if (!queryDetails?.query_sample?.query) {
      return;
    }
    const formatted = formatSQLPlain(queryDetails.query_sample.query);
    if (await copyToClipboard(formatted)) {
      setCopiedSql(true);
      setTimeout(() => setCopiedSql(false), 1800);
    }
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

                <Card className={styles.sectionGap}>
                  <Card.Heading>SQL Query</Card.Heading>
                  <Card.Actions>
                    <Button size="sm" variant="secondary" fill="outline" onClick={handleCopySql}>
                      {copiedSql ? <Check size={13} /> : <Copy size={13} />}
                      <span style={{ marginLeft: 4 }}>{copiedSql ? 'Copied' : 'Copy Formatted SQL'}</span>
                    </Button>
                  </Card.Actions>
                  <Card.Description>
                    <div
                      className={styles.sqlBox}
                      dangerouslySetInnerHTML={{ __html: formatSQL(sample.query) }}
                    />
                  </Card.Description>
                </Card>

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
                <Card.Heading>Executions Over Time</Card.Heading>
                <Card.Description>
                  {chartLoading ? (
                    <LoadingPlaceholder text="Loading execution history..." />
                  ) : chartFrames.length > 0 && chartFrames[0].length > 0 ? (
                    <div style={{ width: '100%' }}>
                      <MyGraph
                        data={chartFrames}
                        loadingState={LoadingState.Done}
                        eventBus={panelEventBus}
                        timeRange={timeRange}
                        onTimeRangeChange={() => {}}
                      />
                    </div>
                  ) : (
                    <Alert title="No Execution History Available" severity="info">
                      No query execution counter time-series data was found for this query hash in the selected time range.
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
