import React, {useEffect, useMemo, useState} from 'react';
import {PanelRenderer} from '@grafana/runtime';
import {AppEvents, DataFrame, EventBusSrv, LoadingState, PanelData, TimeRange} from '@grafana/data';
import {PanelContext, PanelContextProvider} from '@grafana/ui';

export function NestedTablesWithEventBus({
                                      summaryData,
                                      getDetailsData,
                                      timeRange
                                  }: {
    summaryData: DataFrame[];
    getDetailsData: (snapshotId: string) => PanelData;
    timeRange: TimeRange;
}) {
    const [expandedRows, setExpandedRows] = useState<Set<string>>(new Set());
    const [detailsCache, setDetailsCache] = useState<Map<string, PanelData>>(new Map());
    const panelEventBus = useMemo(() => new EventBusSrv(), []);
    const panelData: PanelData = {series: summaryData,
    timeRange: timeRange, state: LoadingState.Loading}
    useEffect(() => {
        // Listen for cell/row click events from the table panel
        const subscription = panelEventBus.subscribe(AppEvents.alertSuccess, (event) => {
            // Table panels emit events when rows are clicked
            console.log('Panel event:', event);
        });

        // Better: Listen for data selection events
        const dataSubscription = panelEventBus.getStream({
            type: 'table-row-click' // Custom event from table interactions
        }).subscribe({
            next: (event: any) => {
                const rowIndex = event.payload?.rowIndex;
                const rowData = event.payload?.row;

                if (rowData && rowData.id) {
                    handleRowToggle(rowData.id);
                }
            }
        });

        return () => {
            subscription.unsubscribe();
            dataSubscription.unsubscribe();
        };
    }, [panelEventBus]);

    const handleRowToggle = (snapshotId: string) => {
        setExpandedRows(prev => {
            const next = new Set(prev);
            if (next.has(snapshotId)) {
                next.delete(snapshotId);
            } else {
                next.add(snapshotId);
                // Fetch details if not cached
                if (!detailsCache.has(snapshotId)) {
                    const details = getDetailsData(snapshotId);
                    setDetailsCache(new Map(detailsCache).set(snapshotId, details));
                }
            }
            return next;
        });
    };

    const panelContext = useMemo(() => ({
        eventBus: panelEventBus,
        onInstanceStateChange: () => {},
        canAddAnnotations: () => false,
        canEditAnnotations: () => false,
        canDeleteAnnotations: () => false,
    }), [panelEventBus]);

    return (
        <div>
            <PanelContextProvider value={panelContext}>
                <PanelRenderer
                    pluginId="table"
                    width={1200}
                    height={400}
                    data={panelData}
                    timeRange={timeRange}
                    // options={{
                    //     showHeader: true,
                    //     cellHeight: 'md',
                    //     footer: {
                    //         show: false,
                    //     },
                    // }}
                />
            </PanelContextProvider>

            {/*/!* Detail Tables *!/*/}
            {/*{Array.from(expandedRows).map(snapshotId => {*/}
            {/*    const detailData = detailsCache.get(snapshotId);*/}
            {/*    if (!detailData) return null;*/}

            {/*    return (*/}
            {/*        <div*/}
            {/*            key={snapshotId}*/}
            {/*            style={{*/}
            {/*                marginLeft: 40,*/}
            {/*                marginTop: 10,*/}
            {/*                marginBottom: 20,*/}
            {/*                border: '1px solid #444',*/}
            {/*                padding: 10,*/}
            {/*                borderRadius: 4*/}
            {/*            }}*/}
            {/*        >*/}
            {/*            <h4>Details for Snapshot: {snapshotId}</h4>*/}
            {/*            <PanelRenderer*/}
            {/*                pluginId="table"*/}
            {/*                width={1100}*/}
            {/*                height={300}*/}
            {/*                data={detailData}*/}
            {/*                timeRange={timeRange}*/}
            {/*            />*/}
            {/*        </div>*/}
            {/*    );*/}
            {/*})}*/}
        </div>
    )
}