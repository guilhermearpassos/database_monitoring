import React, {useCallback, useEffect, useState} from 'react';
import {FetchResponse, getBackendSrv, getDataSourceSrv, PluginPage} from '@grafana/runtime';
import {lastValueFrom, Observable} from 'rxjs';
import {Alert, Button, Card, LegendDisplayMode, LoadingPlaceholder, Select, TimeSeries, useStyles2} from '@grafana/ui';
import {
    CoreApp,
    DataFrame,
    dataFrameFromJSON,
    DataQueryRequest,
    DataSourceApi,
    dateTime,
    FieldType,
    GrafanaTheme2,
    RawTimeRange,
    ScopedVars,
    SelectableValue,
    TimeRange
} from '@grafana/data';
import {css} from '@emotion/css';
import {MyQuery} from '../nested-datasource/types';
import {DataQueryResponse} from "@grafana/data/dist/types/types/datasource";

// Updated interfaces based on your protobuf definitions
interface ServerMetadata {
    name: string;
    type: string;
    host?: string;
}

interface ServerSummary {
    name: string;
    type: string;
    connections: number;
    requestRate: number;
    connectionsByWaitGroup: Record<string, number>;
}

interface DBSnapshot {
    id: string;
    timestamp: string;
    server: ServerMetadata;
    // Add other fields from your snapshot.proto as needed
}

interface QueryMetric {
    sqlHandle: string;
    host: string;
    database: string;
    avgDuration: number;
    executionCount: number;
    // Add other fields from your proto definitions
}

interface SnapshotSummary {
    id: string;
    timestamp: string;
    server: ServerMetadata;
    connectionsByWaitEvent: Record<string, number>;
    timeMsByWaitEvent: Record<string, number>;
}

interface PaginatedResponse<T> {
    data: T[];
    totalCount: number;
    pageNumber: number;
}

// Define table data interface that extends the expected TableData structure
interface SnapshotTableRow {
    id: string;
    timestamp: string;
    servername: string;
    servertype: string;

    [key: string]: any; // This allows for additional properties
}

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

const PageOne = () => {
    const styles = useStyles2(getStyles);

    // State management
    const [servers, setServers] = useState<ServerMetadata[]>([]);
    const [selectedServer, setSelectedServer] = useState<SelectableValue<string> | null>(null);
    const [snapshots, setSnapshots] = useState<DBSnapshot[]>([]);
    const [chartFrames, setChartFrames] = useState<DataFrame[]>([]);
    const [currentPage, setCurrentPage] = useState(1);
    const [totalCount, setTotalCount] = useState(0);
    const [pageSize] = useState(10);
    const [datasource, setDatasource] = useState<DataSourceApi | null>(null);

    // Loading states
    const [serversLoading, setServersLoading] = useState(false);
    const [dataLoading, setDataLoading] = useState(false);
    const [chartLoading, setChartLoading] = useState(false);

    // Error states
    const [error, setError] = useState<string | null>(null);

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
    const loadServers = useCallback(async () => {
        try {
            setServersLoading(true);
            setError(null);

            const now = new Date();
            const oneHourAgo = new Date(now.getTime() - 60 * 60 * 1000);

            const params = {
                start: oneHourAgo.toISOString(),
                end: now.toISOString(),
                type: "databases"
            };

            const response = await makeApiCall<{
                servers: { label: string, value: string }[]
            }>('datasource-options', params);
            let servers: ServerMetadata[] = []
            for (const server of response) {
                servers.push({name: server.label, type: server.label, host: server.label})
            }
            setServers(servers || []);
        } catch (err) {
            setError(err instanceof Error ? err.message : 'Failed to load servers');
        } finally {
            setServersLoading(false);
        }
    }, []);

    // // Load snapshots with pagination
    // const loadSnapshots = useCallback(async (serverName: string, page: number) => {
    //     try {
    //         setDataLoading(true);
    //         setError(null);
    //
    //         const now = new Date();
    //         const oneHourAgo = new Date(now.getTime() - 60 * 60 * 1000);
    //
    //         const params = {
    //             start: oneHourAgo.toISOString(),
    //             end: now.toISOString(),
    //             host: serverName,
    //             page_size: pageSize.toString(),
    //             page_number: page.toString()
    //         };
    //
    //         const response = await makeApiCall<{
    //             snapshots: DBSnapshot[];
    //             totalCount: number;
    //             pageNumber: number;
    //         }>('list-snapshots', params);
    //
    //         setSnapshots(response.snapshots || []);
    //         setTotalCount(response.totalCount || 0);
    //         setCurrentPage(response.pageNumber || page);
    //     } catch (err) {
    //         setError(err instanceof Error ? err.message : 'Failed to load snapshots');
    //     } finally {
    //         setDataLoading(false);
    //     }
    // }, [pageSize]);

    // Load chart data using the datasource
    const loadChartData = useCallback(async (serverName: string) => {
        try {
            setChartLoading(true);

            if (!datasource) {
                console.warn('Datasource not available, falling back to direct API calls');
                // Fallback to the original implementation
                const now = new Date();
                const oneHourAgo = new Date(now.getTime() - 60 * 60 * 1000);

                const params = {
                    start: oneHourAgo.toISOString(),
                    end: now.toISOString(),
                    server: serverName
                };

                const response = await makeApiCall<{
                    snapSummaries: SnapshotSummary[]
                }>('list-snapshot-summaries', params);
                const summaries = response.snapSummaries || [];

                if (summaries.length === 0) {
                    setChartFrames([]);
                    return;
                }

                // Convert backend data to DataFrame format
                const timestamps = summaries.map(s => new Date(s.timestamp).getTime());
                const connectionCounts = summaries.map(s =>
                    Object.values(s.connectionsByWaitEvent).reduce((sum, count) => sum + count, 0)
                );
                const avgWaitTimes = summaries.map(s => {
                    const totalTime = Object.values(s.timeMsByWaitEvent).reduce((sum, time) => sum + time, 0);
                    const totalConnections = Object.values(s.connectionsByWaitEvent).reduce((sum, count) => sum + count, 0);
                    return totalConnections > 0 ? totalTime / totalConnections : 0;
                });

                // In your loadChartData function, replace the DataFrame creation with:
                const frames: DataFrame[] = [{
                    refId: 'database-metrics',
                    name: 'Database Metrics',
                    meta: {},
                    fields: [
                        {
                            name: 'Time',
                            type: FieldType.time,
                            values: new Array(timestamps.length),
                            config: {},
                        },
                        {
                            name: 'Active Connections',
                            type: FieldType.number,
                            values: new Array(connectionCounts.length),
                            config: {
                                displayName: 'Active Connections',
                            }
                        },
                        {
                            name: 'Avg Wait Time (ms)',
                            type: FieldType.number,
                            values: new Array(avgWaitTimes.length),
                            config: {
                                displayName: 'Avg Wait Time (ms)',
                            }
                        }
                    ],
                    length: timestamps.length
                }];

                // Populate the values
                frames[0].fields[0].values = timestamps;
                frames[0].fields[1].values = connectionCounts;
                frames[0].fields[2].values = avgWaitTimes;

                setChartFrames(frames);
                return;
            }

            // Use datasource to query chart data
            const now = Date.now();
            const oneHourAgo = now - (60 * 60 * 1000);

            const timeRange: TimeRange = {
                from: dateTime(oneHourAgo),
                to: dateTime(now),
                raw: {from: 'now-1h', to: 'now'} as RawTimeRange
            };

            const query: MyQuery = {
                refId: 'chart-data',
                database: serverName,
                hide: false,
                key: `chart-${Date.now()}`,
                queryType: '',
                datasource: {
                    type: 'guilhermearpassos-sqlsights-datasource',
                    uid: datasource.uid || ''
                }
            };

            const scopedVars: ScopedVars = {};

            const queryRequest: DataQueryRequest<MyQuery> = {
                app: CoreApp.Dashboard,
                requestId: `chart-${Date.now()}`,
                timezone: 'browser',
                panelId: 1,
                dashboardUID: '',
                range: timeRange,
                timeInfo: '',
                interval: '30s',
                intervalMs: 30000,
                targets: [query],
                maxDataPoints: 100,
                scopedVars: scopedVars,
                startTime: Date.now(),
                liveStreaming: false
            };

            const response: Promise<DataQueryResponse> | Observable<DataQueryResponse> = datasource.query(queryRequest);
            if (response instanceof Promise) {
                response.then(result => processResponse(result));
            } else {
                response.subscribe(result => processResponse(result));
            }

        } catch (err) {
            console.error('Failed to load chart data:', err);
            setChartFrames([]);
        } finally {
            setChartLoading(false);
        }
    }, [datasource]);

    const processResponse = (result: DataQueryResponse) => {
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

        setChartFrames(transformedFrames);
    };

    // Handle server selection
    const handleServerChange = useCallback((option: SelectableValue<string>) => {
        setSelectedServer(option);
        if (option?.value) {
            setCurrentPage(1);
            // loadSnapshots(option.value, 1);
            loadChartData(option.value);
        } else {
            setSnapshots([]);
            setChartFrames([]);
        }
    }, [/*loadSnapshots,*/ loadChartData]);

    // Handle pagination
    const handlePageChange = useCallback((page: number) => {
        if (selectedServer?.value) {
            // loadSnapshots(selectedServer.value, page);
        }
    }, [selectedServer/*, loadSnapshots*/]);

    // Load servers on component mount
    useEffect(() => {
        loadServers();
    }, [loadServers]);

    // Prepare server options for Select component
    const serverOptions: Array<SelectableValue<string>> = servers.map(server => ({
        label: `${server.name} (${server.type})`,
        value: server.name,
        description: server.host
    }));

    // Prepare table columns for InteractiveTable - use the expected format
    const tableColumns = [
        {
            id: 'timestamp',
            header: 'Timestamp',
            sortType: 'string',
            cell: (props: any) => props.value
        },
        {
            id: 'servername',
            header: 'Server',
            sortType: 'string',
            cell: (props: any) => props.value
        },
        {
            id: 'servertype',
            header: 'Type',
            sortType: 'string',
            cell: (props: any) => props.value
        },
        {
            id: 'id',
            header: 'Snapshot ID',
            sortType: 'string',
            cell: (props: any) => props.value
        }
    ];

    // Prepare table data from snapshots - flatten the structure
    const tableData = snapshots.map(snapshot => ({
        timestamp: new Date(snapshot.timestamp).toLocaleString(),
        servername: snapshot.server?.name || 'Unknown',
        servertype: snapshot.server?.type || 'Unknown',
        id: snapshot.id
    }));

    return (
        <PluginPage>
            <div className={styles.container}>
                <h2>SQL Database Monitoring</h2>
                <p>Monitor and analyze database performance and activity across your servers.</p>

                {error && (
                    <Alert title="Error" severity="error">
                        {error}
                    </Alert>
                )}

                <div className={styles.section}>
                    <h3>Server Selection</h3>
                    <div className={styles.serverSelection}>
                        <Select
                            placeholder="Select a server..."
                            options={serverOptions}
                            value={selectedServer}
                            onChange={handleServerChange}
                            isLoading={serversLoading}
                            isClearable
                        />
                    </div>
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
                                        <TimeSeries
                                            frames={chartFrames}
                                            timeRange={{
                                                from: dateTime(Math.min(...(chartFrames[0].fields[0].values as number[]))),
                                                to: dateTime(Math.max(...(chartFrames[0].fields[0].values as number[]))),
                                                raw: {from: 'now-1h', to: 'now'}
                                            }}
                                            width={800}
                                            height={350}
                                            legend={{
                                                calcs: new Array<string>(),
                                                displayMode: LegendDisplayMode.List,
                                                showLegend: false, placement: "right"
                                            }}
                                            timeZone={"utc"}
                                        />
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
                                <div className={styles.tableContainer}>
                                    {dataLoading ? (
                                        <LoadingPlaceholder text="Loading snapshots..."/>
                                    ) : (
                                        <>
                                            {/*<InteractiveTable*/}
                                            {/*    columns={tableColumns}*/}
                                            {/*    data={tableData as any[]}*/}
                                            {/*    getRowId={(row) => row.id}*/}
                                            {/*/>*/}

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
                            </Card>
                        </div>
                    </>
                )}
            </div>
        </PluginPage>
    );
};

export default PageOne;