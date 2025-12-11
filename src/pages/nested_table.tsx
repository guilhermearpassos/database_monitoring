import React, {useEffect, useMemo, useState} from 'react';
import {PanelRenderer} from '@grafana/runtime';
import {
    BusEventWithPayload,
    DataFrame,
    EventBusSrv,
    LoadingState,
    PanelData,
    TimeRange,
} from '@grafana/data';
import {PanelContext, PanelContextProvider,} from '@grafana/ui';

// Define a proper event class
class TableRowClickEvent extends BusEventWithPayload<{ id: string; rowIndex: number }> {
    static type = 'table-row-click';  // Must be static!
}


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
    const panelData: PanelData = {
        series: summaryData,
        timeRange: timeRange, state: LoadingState.Loading
    }

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

    useEffect(() => {

        // Better: Listen for data selection events
        const dataSubscription = panelEventBus.getStream(TableRowClickEvent).subscribe({
            next: (event: any) => {
                const rowID = event.payload?.id;

                if (rowID) {
                    handleRowToggle(rowID);
                }
            }
        });

        return () => {
            dataSubscription.unsubscribe();
        };
    });

    const panelContext: PanelContext = useMemo(() => ({
        eventBus: panelEventBus,
        eventsScope: "sqlsights-one",
        onInstanceStateChange: () => {
        },
        canAddAnnotations: () => false,
        canEditAnnotations: () => false,
        canDeleteAnnotations: () => false,
    }), [panelEventBus]);
    const summaryDataWithLinks: PanelData = {
        ...panelData,
        series: panelData.series.map(series => ({
            ...series,
            fields: series.fields.map(field => {
                if (field.name === 'id' || field.name === 'Database') {
                    return {
                        ...field,
                        config: {
                            ...field.config,
                            links: [{
                                title: 'View Details',
                                url: 'app://expand-row?id=${__data.fields.id}', // Custom scheme
                                targetBlank: false,
                                onClick: event => {
                                    const snapID: string = event.origin.field.values[event.origin.rowIndex];
                                    panelEventBus.publish(
                                        new TableRowClickEvent({
                                            id: snapID,
                                            rowIndex: event.origin.rowIndex
                                        }));

                                }
                            }]
                        }
                    };
                }
                return field;
            })
        }))
    };

    return (
        <div>
            {/*<DataLinksContext value={{dataLinkPostProcessor}}>*/}
            <PanelContextProvider value={panelContext}>
                <PanelRenderer
                    title="nestedtable"
                    pluginId="table"
                    width={1200}
                    height={400}
                    data={summaryDataWithLinks}
                    // options={{
                    //     showHeader: true,
                    //     cellHeight: 'md',
                    //     footer: {
                    //         show: false,
                    //     },
                    // }}
                />
            </PanelContextProvider>
            {/*</DataLinksContext>*/}
            Detail Tables
            {Array.from(expandedRows).map(snapshotId => {
                const detailData = detailsCache.get(snapshotId);
                if (!detailData) {
                    return null;
                }

                return (
                    <div
                        key={snapshotId}
                        style={{
                            marginLeft: 40,
                            marginTop: 10,
                            marginBottom: 20,
                            border: '1px solid #444',
                            padding: 10,
                            borderRadius: 4
                        }}
                    >
                        <h4>Details for Snapshot: {snapshotId}</h4>
                        <PanelRenderer
                            title="detailsnestedtable"
                            pluginId="table"
                            width={1100}
                            height={300}
                            data={detailData}
                        />
                    </div>
                );
            })}
        </div>
    )
}
