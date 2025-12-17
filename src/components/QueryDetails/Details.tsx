import React, {useCallback, useEffect, useMemo, useState} from "react";
import {ExecutionPlanViewer} from "../ExecutionPlanTree/plan";
import {Alert, Card} from "@grafana/ui";
import {lastValueFrom, Observable} from "rxjs";
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


export const QueryDetailsComponent: React.FC<{ sampleID: string, snapID: string, timeRange: TimeRange, server: string}> = ({sampleID, snapID, timeRange, server}) => {

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
                const ds = await getDataSourceSrv().get({type:'guilhermearpassos-sqlsights-datasource'});
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
        if (!queryDetails){
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
            {queryDetails && (<MyGraph
                data={chartFrames}
                loadingState={chartLoading ? LoadingState.Loading : LoadingState.Done}
                eventBus={panelEventBus}
                timeRange={timeRange}
                onTimeRangeChange={range => {}}
                width={800}
                height={400}
            />)}
            {queryDetails && queryDetails.blocking_chain && queryDetails.blocking_chain.roots.length > 0 && (<BlockingChainComponent chain={queryDetails.blocking_chain} currentSampleId={sample}/>)}
            {queryDetails && queryDetails.plan && (<ExecutionPlanViewer executionPlan={queryDetails.plan}/>)}
        </Card.Description>
    </Card>)
}
