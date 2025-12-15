import React, {useCallback, useEffect, useMemo, useState} from 'react';
import {FetchResponse, getBackendSrv, getDataSourceSrv, PluginPage} from '@grafana/runtime';
import {lastValueFrom, Observable} from 'rxjs';
import {
    Alert,
    Button,
    Card,
    Combobox,
    ComboboxOption,
    LoadingPlaceholder,
    TimeRangePicker,
    useStyles2
} from '@grafana/ui';
import {
    CoreApp,
    DataFrame,
    dataFrameFromJSON,
    DataQueryRequest,
    DataSourceApi,
    dateTime,
    EventBusSrv,
    GrafanaTheme2,
    LoadingState, PanelData,
    TimeRange
} from '@grafana/data';
import {css} from '@emotion/css';
import {MyQuery} from '../nested-datasource/types';
import {DataQueryResponse} from "@grafana/data/dist/types/types/datasource";
import {MyGraph} from "./graph";
import {NestedTablesWithEventBus} from "./nested_table";
import {addWaitTypeHTMLColumn} from "./waitTypeField";
import {ExecutionPlanViewer, ParsedExecutionPlan} from "../components/ExecutionPlanTree/plan";
// ... existing code ...

const getStyles = (theme: GrafanaTheme2) => ({
    container: css`
        padding: ${theme.spacing(2)};
    `,
    section: css`
        margin-bottom: ${theme.spacing(3)};
    `,
    serverSelection: css`
        max-width: 300px;
        margin-bottom: ${theme.spacing(2)};
    `,
    headerRow: css`
        display: flex;
        align-items: flex-start;
        justify-content: space-between;
        gap: ${theme.spacing(2)};
        flex-wrap: wrap;
        margin-bottom: ${theme.spacing(2)};
    `,
    headerLeft: css`
        min-width: 280px;
    `,
    headerRight: css`
        margin-left: auto;
        display: flex;
        justify-content: flex-end;
        align-items: center;
    `,
    timeRangeLabel: css`
        font-weight: ${theme.typography.fontWeightMedium};
    `,
    chartContainer: css`
        height: 400px;
        margin-bottom: ${theme.spacing(3)};
        border: 1px solid ${theme.colors.border.medium};
        border-radius: ${theme.shape.borderRadius()};
        padding: ${theme.spacing(2)};
    `,
    tableContainer: css`
        margin-top: ${theme.spacing(2)};
    `,
    paginationContainer: css`
        display: flex;
        justify-content: space-between;
        align-items: center;
        margin-top: ${theme.spacing(2)};
        padding: ${theme.spacing(1)};
        border-top: 1px solid ${theme.colors.border.medium};
    `,
    loadingContainer: css`
        height: 200px;
        display: flex;
        align-items: center;
        justify-content: center;
    `
});

interface ServerMetadata {
    name: string;
    type: string;
    host?: string;
}
interface sampleID {
    sampleID: string;
    snapId: string;
}

const PageOne = () => {
    const styles = useStyles2(getStyles);

    const panelEventBus = useMemo(() => new EventBusSrv(), []);

    // State management
    const [servers, setServers] = useState<ServerMetadata[]>([]);
    const [selectedServer, setSelectedServer] = useState<ComboboxOption | null>(null);
    const [snapshots, setSnapshots] = useState<DataFrame[]>([]);
    const [chartFrames, setChartFrames] = useState<DataFrame[]>([]);
    const [currentPage, setCurrentPage] = useState(1);
    const [totalCount, setTotalCount] = useState(0);
    const [pageSize] = useState(10);
    const [datasource, setDatasource] = useState<DataSourceApi | null>(null);
    const [sampleID, setSampleID] = useState<sampleID | null>(null);
    const [executionPlan, setExecutionPlan] = useState<ParsedExecutionPlan | null>(null);
    const [chartTimeRange, setChartTimeRange] = useState<TimeRange>(() => ({
        from: dateTime().subtract(1, 'hour'),
        to: dateTime(),
        raw: {from: 'now-1h', to: 'now'},
    }));

    // Loading states
    // const [serversLoading, setServersLoading] = useState(false);
    const [dataLoading, setDataLoading] = useState(false);
    const [chartLoading, setChartLoading] = useState(false);

    // Error states
    const [error, setError] = useState<string | null>(null);

    // Function to fetch HTML content from backend
    useEffect(() => {
        const fetchPlan = async () => {
            if (!sampleID){
                return;
            }
            try {
                let response: Observable<FetchResponse<string>>;
                response = await getBackendSrv().fetch({
                    url: '/api/plugins/guilhermearpassos-sqlsights-app/resources/getExecPlan?sampleId='+sampleID?.sampleID+'&snapId='+sampleID?.snapId,
                });
                // Get the response as text since it's HTML
                const textResponse: { data: string } = await lastValueFrom(response);
                setExecutionPlan(textResponse.data);
            } catch (err) {
                throw new Error(`Failed to fetch: ${err}`);
            }
        };
        fetchPlan()
    }, [sampleID]);

    // Initialize datasource
    useEffect(() => {
        const initDatasource = async () => {
            try {
                // Get the first available datasource of our type
                const ds = await getDataSourceSrv().get('guilhermearpassos-sqlsights-datasource');
                setDatasource(ds);
            } catch (err) {
                console.error('Failed to initialize datasource:', err);
                // Fallback to direct API calls if datasource is not available
            }
        };
        initDatasource();
    }, []);

    // Generic API call function
    const makeApiCall = async <T, >(endpoint: string, params?: Record<string, any>): Promise<T> => {
        try {
            const url = `/api/plugins/guilhermearpassos-sqlsights-app/resources/${endpoint}`;
            const queryParams = params ? '?' + new URLSearchParams(params).toString() : '';

            const response: Observable<FetchResponse<T>> = getBackendSrv().fetch({
                url: url + queryParams,
            });

            const result = await lastValueFrom(response);
            return result.data;
        } catch (err) {
            throw new Error(`Failed to fetch ${endpoint}: ${err}`);
        }
    };

    // Load available servers from backend
    const loadServers = useCallback(async (timeRange: TimeRange) => {
        try {
            // setServersLoading(true);
            setError(null);

            const params = {
                start: timeRange.from.toISOString(),
                end: timeRange.to.toISOString(),
                type: "databases"
            };

            const response = await makeApiCall<Array<{ label: string, value: string }>>('datasource-options', params);
            let servers: ServerMetadata[] = []
            for (const server of response) {
                servers.push({name: server.label, type: server.label, host: server.label})
            }
            setServers(servers || []);
        } catch (err) {
            setError(err instanceof Error ? err.message : 'Failed to load servers');
        } finally {
            // setServersLoading(false);
        }
    }, []);
    // Load servers on component mount
    useEffect(() => {
        loadServers(chartTimeRange);
    }, [chartTimeRange, loadServers]);

    // // Load snapshots with pagination
    const loadSnapshots = useCallback(async (serverName: string, page: number, timeRange: TimeRange) => {
        if (!datasource) {
            console.warn('Datasource not available');
            setSnapshots([]);
            return;
        }
        try {
            setDataLoading(true);
            setError(null);

            console.log(pageSize)
            console.log(page)
            // This mirrors what Explore does - simple query structure
            const query: MyQuery = {
                refId: 'A', // Explore always starts with 'A'
                database: serverName,
                hide: false,
                datasource: {
                    type: datasource.type,
                    uid: datasource.uid
                },
                queryType: "snapshot-list"
            };

            // Mimic Explore's query request structure exactly
            const queryRequest: DataQueryRequest<MyQuery> = {
                app: CoreApp.Explore, // Use Explore app context
                requestId: `explore_${Date.now()}`,
                timezone: 'browser',
                panelId: 1,
                dashboardUID: '',
                range: timeRange,
                timeInfo: '',
                interval: '1m',
                intervalMs: 60000,
                targets: [query],
                maxDataPoints: 1000, // Explore uses higher maxDataPoints
                scopedVars: {},
                startTime: Date.now(),
                liveStreaming: false
            };

            console.log('Query request (same as Explore):', queryRequest);

            setTotalCount(1)
            // Use the datasource query method directly - exactly like Explore
            const resultObservable = datasource.query(queryRequest);
            let result: DataQueryResponse
            if (resultObservable instanceof Observable) {
                result = await lastValueFrom(resultObservable);

            } else {
                result = await resultObservable;
            }
            console.log('Query response (raw):', result);

            // Explore doesn't transform the data - it uses it directly
            if (result.data && Array.isArray(result.data)) {
                const transformedFrames: DataFrame[] = processResponse(result)
                setSnapshots(transformedFrames.map((f) => addWaitTypeHTMLColumn(f, "waitsByType", "Waits")));
            } else {
                setSnapshots([]);
            }
        } catch (err) {
            setError(err instanceof Error ? err.message : 'Failed to load snapshots');
        } finally {
            setDataLoading(false);
        }
    }, [datasource, pageSize]);


    // Load chart data using the datasource
    const loadChartData = useCallback(async (serverName: string, timeRange: TimeRange) => {
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
                database: serverName,
                hide: false,
                datasource: {
                    type: datasource.type,
                    uid: datasource.uid
                },
                queryType: "chart"
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

    // Handle server selection
    const handleServerChange = useCallback((option: ComboboxOption<string> | null) => {
        setSelectedServer(option);
        setCurrentPage(1);

    }, []);

    const handleChartTimeRangeChange = useCallback((nextRange: TimeRange) => {
        setChartTimeRange(nextRange);
        setCurrentPage(1);
    }, []);

    useEffect(() => {
        const serverName = selectedServer?.value;
        if (!serverName) {
            return;
        }
        const tr: TimeRange = {from: chartTimeRange.from, to: chartTimeRange.to, raw: chartTimeRange.raw};
        loadChartData(serverName, tr);
        loadSnapshots(serverName, 1, tr);
    }, [
        selectedServer?.value,
        chartTimeRange.from,
        chartTimeRange.to,
        chartTimeRange.raw,
        loadChartData,
        loadSnapshots
    ]);

    // Handle pagination
    const handlePageChange = useCallback((page: number) => {
        if (selectedServer?.value) {
            // loadSnapshots(selectedServer.value, page);
        }
    }, [selectedServer/*, loadSnapshots*/]);


    // Prepare server options for Select component
    const serverOptions: ComboboxOption[] = servers.map(server => ({
        label: `${server.name} (${server.type})`,
        value: server.name,
        description: server.host
    }));


    const loadSamplesPanelData = useCallback(async (snapID: string): Promise<DataFrame[]> => {
        if (!datasource) {
            throw new Error('Datasource not available');
        }
        const serverName = selectedServer?.value;
        if (!serverName) {
            return [];
        }
        if (!snapID) {
            return [];
        }

        const query: MyQuery = {
            refId: 'A',
            database: serverName,
            hide: false,
            datasource: {
                type: datasource.type,
                uid: datasource.uid
            },
            queryType: "snapshot",
            snapshotID: snapID
        };

        const queryRequest: DataQueryRequest<MyQuery> = {
            app: CoreApp.Explore,
            requestId: `explore_${Date.now()}`,
            timezone: 'browser',
            panelId: 1,
            dashboardUID: '',
            range: chartTimeRange,
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
        const result: DataQueryResponse =
            resultObservable instanceof Observable ? await lastValueFrom(resultObservable) : await resultObservable;

        return processResponse(result);
    }, [datasource, selectedServer?.value, chartTimeRange]);

    const onSampleSelection = useCallback((snapID: string, sampleID: string) => {setSampleID({sampleID: sampleID, snapId: snapID})}, [])
    const getDetailsData = useCallback(async (id: string): Promise<PanelData> => {
        let frames = await loadSamplesPanelData(id);
        return {

            series: frames, state: LoadingState.Loading,
            timeRange: {
                from: dateTime().subtract(1, 'hour'),
                to: dateTime(),
                raw: {from: 'now-1h', to: 'now'}
            }
        };
    }, [loadSamplesPanelData]);
    return (
        <PluginPage>
            <div className={styles.headerRow}>
                <div className={styles.headerLeft}>
                    <h2>SQL Database Monitoring</h2>
                    <p>Monitor and analyze database performance and activity across your servers.</p>
                </div>

                <div className={styles.headerRight}>
                    <TimeRangePicker
                        value={chartTimeRange}
                        onChange={handleChartTimeRangeChange}
                        timeZone="browser"
                        onChangeTimeZone={timeZone => {
                        }}
                        onZoom={() => {
                        }}
                        onMoveBackward={() => {
                        }}
                        onMoveForward={() => {
                        }}/>
                </div>
                {error && (
                    <Alert title="Error" severity="error">
                        {error}
                    </Alert>
                )}
            </div>


            <div className={styles.section}>
                <h3>Server Selection</h3>
                <div className={styles.serverSelection}>
                    <Combobox
                        width="auto"
                        options={serverOptions}
                        value={selectedServer}
                        onChange={handleServerChange}
                        placeholder="Select a server..."
                        minWidth={20}
                        isClearable
                    />
                </div>

                {selectedServer && (
                    <>
                        <div className={styles.section}>
                            <Card>
                                <Card.Heading>Database Activity Over Time</Card.Heading>
                                <div className={styles.chartContainer}>
                                    {chartLoading ? (
                                        <div className={styles.loadingContainer}>
                                            <LoadingPlaceholder text="Loading chart data..."/>
                                        </div>
                                    ) : chartFrames.length > 0 ? (
                                        <>
                                            {chartFrames.length > 0 && (
                                                <MyGraph
                                                    data={chartFrames}
                                                    loadingState={chartLoading ? LoadingState.Loading : LoadingState.Done}
                                                    eventBus={panelEventBus}
                                                    timeRange={chartTimeRange}
                                                    onTimeRangeChange={handleChartTimeRangeChange}
                                                    width={800}
                                                    height={400}
                                                />
                                            )}
                                        </>
                                    ) : (
                                        <div className={styles.loadingContainer}>
                                            <p>No chart data available</p>
                                        </div>
                                    )}
                                </div>
                            </Card>
                        </div>

                        <div className={styles.section}>
                            <Card>
                                <Card.Heading>Database Snapshots</Card.Heading>
                                <Card.Description>

                                <div className={styles.tableContainer}>
                                    {dataLoading ? (
                                        <LoadingPlaceholder text="Loading snapshots..."/>
                                    ) : (
                                        <>
                                            <NestedTablesWithEventBus
                                                getDetailsData={
                                                    getDetailsData
                                                }
                                                summaryData={snapshots}
                                                timeRange={{
                                                    from: dateTime().subtract(1, 'hour'),
                                                    to: dateTime(),
                                                    raw: {from: 'now-1h', to: 'now'}
                                                }}
                                                onSampleSelection={onSampleSelection}
                                            />

                                            <div className={styles.paginationContainer}>
                        <span>
                          Showing {((currentPage - 1) * pageSize) + 1} to {Math.min(currentPage * pageSize, totalCount)} of {totalCount} entries
                        </span>
                                                <div>
                                                    <Button
                                                        variant="secondary"
                                                        onClick={() => handlePageChange(currentPage - 1)}
                                                        disabled={currentPage <= 1}
                                                        style={{marginRight: '8px'}}
                                                    >
                                                        Previous
                                                    </Button>
                                                    <span style={{margin: '0 16px'}}>
                            Page {currentPage} of {Math.ceil(totalCount / pageSize)}
                          </span>
                                                    <Button
                                                        variant="secondary"
                                                        onClick={() => handlePageChange(currentPage + 1)}
                                                        disabled={currentPage >= Math.ceil(totalCount / pageSize)}
                                                    >
                                                        Next
                                                    </Button>
                                                </div>
                                            </div>
                                        </>
                                    )}
                                </div>
                                </Card.Description>
                            </Card>
                        </div>
                        {executionPlan && (
                            <div className={styles.section}>

                                <Card>
                                    <Card.Heading>Database Snapshots</Card.Heading>
                                    <Card.Description>
                                        <ExecutionPlanViewer executionPlan={executionPlan}/>
                                    </Card.Description>
                                </Card>
                            </div>
                        )}
                    </>
                )}
            </div>
        </PluginPage>
    );
};

export default PageOne;
