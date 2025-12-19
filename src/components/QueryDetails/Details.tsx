import React, {useCallback, useEffect, useMemo, useState} from "react";
import {ExecutionPlanViewer} from "../ExecutionPlanTree/plan";
import {Alert, Card, Grid, Tooltip, useTheme2} from "@grafana/ui";
import {lastValueFrom, Observable} from "rxjs";
import Prism from 'prismjs';
import {FetchResponse, getBackendSrv, getDataSourceSrv} from "@grafana/runtime";
import {BlockingChainComponent} from "./blockingChain";
import {QueryDetails} from "./types";
import {
    CoreApp,
    DataFrame,
    dataFrameFromJSON,
    DataQueryRequest,
    DataSourceApi, EventBusSrv,
    LoadingState,
    TimeRange
} from "@grafana/data";
import {MyQuery} from "../../nested-datasource/types";
import {DataQueryResponse} from "@grafana/data/dist/types/types/datasource";
import {MyGraph} from "../../pages/graph";
import {formatSQL} from "../../utils/formatters";
import {getStyles} from "../../utils/styles";


export const QueryDetailsComponent: React.FC<{
    sampleID: string,
    snapID: string,
    timeRange: TimeRange,
    server: string
}> = ({sampleID, snapID, timeRange, server}) => {
    const theme = useTheme2();
    const styles = getStyles(theme);

    const panelEventBus = useMemo(() => new EventBusSrv(), []);
    const [queryDetails, setQueryDetails] = useState<QueryDetails | null>(null);
    const [sample] = useState<string>(sampleID);
    const [snap] = useState<string>(snapID);

    const [chartFrames, setChartFrames] = useState<DataFrame[]>([]);
    const [chartLoading, setChartLoading] = useState(false);
    const [error, setError] = useState<string | null>(null);

    const [datasource, setDatasource] = useState<DataSourceApi | null>(null);

    // Initialize datasource
    useEffect(() => {
        const initDatasource = async () => {
            try {
                // Get the first available datasource of our type
                const ds = await getDataSourceSrv().get({type: 'guilhermearpassos-sqlsights-datasource'});
                setDatasource(ds);
            } catch (err) {
                console.error('Failed to initialize datasource:', err);
                // Fallback to direct API calls if datasource is not available
            }
        };
        initDatasource();
    }, []);

    const processResponse = (result: DataQueryResponse): DataFrame[] => {
        const transformedFrames: DataFrame[] = [];

        result.data.forEach((rawFrame: any) => {
            try {
                // Try to use Grafana's utility to convert from JSON
                const frame = dataFrameFromJSON({
                    schema: rawFrame.schema,
                    data: rawFrame.data
                });
                transformedFrames.push(frame);
            } catch (error) {
                console.error('Failed to convert frame:', error);
                // Fallback to manual transformation if needed
            }
        });

        return transformedFrames
    };
    // Load chart data using the datasource
    const loadChartData = useCallback(async (qd: QueryDetails, timeRange: TimeRange, server: string) => {
        if (!datasource) {
            console.warn('Datasource not available');
            setChartFrames([]);
            return;
        }

        try {
            setChartLoading(true);
            setError(null);

            const query: MyQuery = {
                refId: 'A',
                database: server,
                hide: false,
                datasource: {
                    type: datasource.type,
                    uid: datasource.uid
                },
                queryType: "metrics_series",
                queryHash: qd.query_sample.query_hash,
                metrics: ["executionCount"]
            };

            const queryRequest: DataQueryRequest<MyQuery> = {
                app: CoreApp.Explore,
                requestId: `explore_${Date.now()}`,
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
                liveStreaming: false
            };

            const resultObservable = datasource.query(queryRequest);
            let result: DataQueryResponse;
            if (resultObservable instanceof Observable) {
                result = await lastValueFrom(resultObservable);
            } else {
                result = await resultObservable;
            }

            if (result.data && Array.isArray(result.data)) {
                const transformedFrames: DataFrame[] = processResponse(result);
                setChartFrames(transformedFrames);
            } else {
                setChartFrames([]);
            }
        } catch (err) {
            console.error('Query failed:', err);
            setError(err instanceof Error ? err.message : 'Query failed');
            setChartFrames([]);
        } finally {
            setChartLoading(false);
        }
    }, [datasource]);

    useEffect(() => {
        if (!queryDetails) {
            return;
        }
        loadChartData(queryDetails, timeRange, server)
    }, [queryDetails, timeRange, server, loadChartData]);
    // Function to fetch HTML content from backend
    useEffect(() => {
        const fetchPlan = async () => {
            if (!sample) {
                return;
            }
            try {
                let response: Observable<FetchResponse<QueryDetails>>;
                response = await getBackendSrv().fetch({
                    url: '/api/plugins/guilhermearpassos-sqlsights-app/resources/getQueryDetails?sampleId=' + sample + '&snapId=' + snap,
                });
                // Get the response as text since it's HTML
                const textResponse: { data: QueryDetails } = await lastValueFrom(response);
                setQueryDetails(textResponse.data);
            } catch (err) {
                throw new Error(`Failed to fetch: ${err}`);
            }
        };
        fetchPlan()
    }, [sample, snap]);
    return (<Card>
        <Card.Heading>Sample Details</Card.Heading>
        <Card.Description>
            {error && (
                <Alert title="Error" severity="error">
                    {error}
                </Alert>
            )}
            {queryDetails && (
                <div>
                    <Card>
                        <Card.Heading>Query Information</Card.Heading>
                        <Card.Description>
                            <Grid columns={3}>
                                <div>
                                    <p className={styles.label}>Status</p>
                                    <p className={styles.sessionId}>{queryDetails.query_sample.status}</p>
                                </div>
                                <div>
                                    <p className={styles.label}>Time Elapsed</p>
                                    <p className={styles.sessionId}>{queryDetails.query_sample.execution_time}</p>
                                </div>
                                <div>
                                    <p className={styles.label}>Query Hash</p>
                                    <p className={styles.sessionId}>{queryDetails.query_sample.query_hash}</p>
                                </div>
                                <div>
                                    <p className={styles.label}>Database</p>
                                    <p className={styles.sessionId}>{queryDetails.query_sample.database}</p>
                                </div>
                            </Grid>
                            <Grid columns={1}>

                                <div>
                                    <p className={styles.label}>Snapshot</p>
                                    <p className={styles.sessionId}>{queryDetails.query_sample.snap_id}</p>
                                </div>
                                <div>
                                    <p className={styles.label}>SQL Handle</p>
                                    <p className={styles.sessionId}>{queryDetails.query_sample.sql_handle}</p>
                                </div>
                                <div>
                                    <p className={styles.label}>Sample</p>
                                    <p className={styles.sessionId}>{queryDetails.query_sample.sample_id}</p>
                                </div>
                                <div>
                                    <p className={styles.label}>SQL Text</p>
                                    <Tooltip interactive={true} content={(
                                        <div className={styles.tooltipContent}>
                                            <div className={styles.formattedSQL}
                                                 dangerouslySetInnerHTML={{__html: formatSQL(queryDetails.query_sample.query)}}>
                                            </div>
                                        </div>)} placement="right">
                                <pre className={styles.singleLine}
                                     dangerouslySetInnerHTML={{__html: Prism.highlight(queryDetails.query_sample.query, Prism.languages.sql, 'sql')}}>

                                </pre>
                                    </Tooltip>
                                </div>
                            </Grid>

                        </Card.Description>
                    </Card>
                    <br/>
                    <Card>
                        <Card.Heading>Session Information</Card.Heading>
                        <Card.Description>
                            <Grid columns={3}>
                                <div>
                                    <p className={styles.label}>Session ID</p>
                                    <p className={styles.sessionId}>{queryDetails.query_sample.sid}</p>
                                </div>
                                <div>
                                    <p className={styles.label}>Session Status</p>
                                    <p className={styles.sessionId}>{queryDetails.query_sample.session_status}</p>
                                </div>
                                <div>
                                    <p className={styles.label}>User Name</p>
                                    <p className={styles.sessionId}>{queryDetails.query_sample.user}</p>
                                </div>
                                <div>
                                    <p className={styles.label}>Host</p>
                                    <p className={styles.sessionId}>{queryDetails.query_sample.session_host}</p>
                                </div>
                                <div>
                                    <p className={styles.label}>Ip</p>
                                    <p className={styles.sessionId}>{queryDetails.query_sample.session_client_ip}</p>
                                </div>
                                <div>
                                    <p className={styles.label}>Program Name</p>
                                    <p className={styles.sessionId}>{queryDetails.query_sample.session_program_name}</p>
                                </div>
                                <div>
                                    <p className={styles.label}>Session Login Time</p>
                                    <p className={styles.sessionId}>{queryDetails.query_sample.session_login_time}</p>
                                </div>
                                <div>
                                    <p className={styles.label}>Last Request Start</p>
                                    <p className={styles.sessionId}>{queryDetails.query_sample.session_last_request_start}</p>
                                </div>
                                <div>
                                    <p className={styles.label}>Last Request End</p>
                                    <p className={styles.sessionId}>{queryDetails.query_sample.session_last_request_end}</p>
                                </div>
                            </Grid>
                        </Card.Description>
                    </Card>
                </div>
            )}
            {chartFrames && chartFrames.length && chartFrames[0].length && (<Card>
                <Card.Heading>Executions over time</Card.Heading>
                <Card.Description>
                    <div style={{maxWidth: '50vw'}}>
                        <MyGraph
                            data={chartFrames}
                            loadingState={chartLoading ? LoadingState.Loading : LoadingState.Done}
                            eventBus={panelEventBus}
                            timeRange={timeRange}
                            onTimeRangeChange={range => {
                            }}
                        />
                    </div>
                </Card.Description>
            </Card>)}
            {queryDetails && queryDetails.blocking_chain && (queryDetails.blocking_chain.roots.length > 0) && (
                <Card>
                    <Card.Heading>Blocking Chain</Card.Heading>
                    <Card.Description>
                        <BlockingChainComponent
                            chain={queryDetails.blocking_chain}
                            currentSampleId={sample}
                        />
                    </Card.Description>
                </Card>
            )}
            {queryDetails && queryDetails.plan && (
                <Card>
                    <Card.Heading>Execution Plan</Card.Heading>
                    <Card.Description>
                        <ExecutionPlanViewer executionPlan={queryDetails.plan}/>
                    </Card.Description>
                </Card>
            )}
        </Card.Description>
    </Card>)
}
