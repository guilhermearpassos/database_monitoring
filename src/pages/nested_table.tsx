import React, {useCallback, useEffect, useMemo, useState} from 'react';
import {PanelRenderer} from '@grafana/runtime';
import {
    BusEventWithPayload,
    DataFrame,
    DataTransformerConfig,
    EventBusSrv, FieldConfigSource,
    FieldType,
    LoadingState,
    MutableDataFrame,
    PanelData,
    TimeRange,
    transformDataFrame,
} from '@grafana/data';
import {PanelContext, PanelContextProvider} from '@grafana/ui';
import {Observable} from 'rxjs';


// Define a proper event class
class TableRowClickEvent extends BusEventWithPayload<{ id: string; rowIndex: number }> {
    static type = 'table-row-click';
}

export function NestedTablesWithEventBus({
                                             summaryData,
                                             getDetailsData,
                                             timeRange,
                                         }: {
    summaryData: DataFrame[];
    getDetailsData: (snapshotId: string) => Promise<PanelData>;
    timeRange: TimeRange;
}) {
    const [expandedRows, setExpandedRows] = useState<Set<string>>(new Set());
    const [detailsCache, setDetailsCache] = useState<Map<string, DataFrame[]>>(new Map());
    const [detailsLoading, setDetailsLoading] = useState<Set<string>>(new Set());
    const [transformedFrames, setTransformedFrames] = useState<DataFrame[]>([]);

    // Track row index to snapshot ID mapping for DOM click handling
    const rowIndexToIdRef = React.useRef<Map<number, string>>(new Map());

    const panelEventBus = useMemo(() => new EventBusSrv(), []);

    const panelContext: PanelContext = useMemo(
        () => ({
            eventBus: panelEventBus,
            eventsScope: 'sqlsights-one',
            onInstanceStateChange: () => {
            },
            canAddAnnotations: () => false,
            canEditAnnotations: () => false,
            canDeleteAnnotations: () => false,
        }),
        [panelEventBus]
    );

    const summaryFrame: DataFrame = useMemo(() => {
        return summaryData ? (summaryData[0]) : {fields: [], length: 0}
    }, [summaryData]);

    const idFieldIndex = useMemo(() => {
        if (!summaryFrame) {
            return -1;
        }
        return summaryFrame.fields.findIndex((f) => f.name === 'id');
    }, [summaryFrame]);

    const fetchDetailsIfNeeded = useCallback(async (snapshotId: string) => {
        if (detailsCache.has(snapshotId) || detailsLoading.has(snapshotId)) {
            return;
        }

        setDetailsLoading((prev) => new Set(prev).add(snapshotId));

        try {
            const details = await getDetailsData(snapshotId);
            setDetailsCache((prev) => {
                const next = new Map(prev);
                next.set(snapshotId, details.series);
                return next;
            });
        } catch (err) {
            console.error('Failed to fetch details:', err);
        } finally {
            setDetailsLoading((prev) => {
                const next = new Set(prev);
                next.delete(snapshotId);
                return next;
            });
        }
    }, [detailsCache, detailsLoading, getDetailsData]);

    const handleRowToggle = useCallback((snapshotId: string) => {
        setExpandedRows((prev) => {
            const next = new Set(prev);
            if (!next.has(snapshotId)) {
                next.add(snapshotId);
                void fetchDetailsIfNeeded(snapshotId);
            }
            return next;
        });
    }, [setExpandedRows, fetchDetailsIfNeeded])

    useEffect(() => {
        const sub = panelEventBus.getStream(TableRowClickEvent).subscribe({
            next: (event: any) => {
                const rowID = event.payload?.id;
                if (rowID) {
                    handleRowToggle(rowID);
                }
            },
        });

        return () => sub.unsubscribe();
    }, [panelEventBus, handleRowToggle]);

    // Observe DOM for nested table expand/collapse buttons
    useEffect(() => {
        const handleExpandButtonClick = (event: Event) => {
            const target = event.target as HTMLElement;
            const button = target.closest('[role="button"]');

            if (!button) {
                return;
            }

            // Check if this is an expand/collapse button (has SVG icon)
            const svg = button.querySelector('svg');
            if (!svg) {
                return;
            }

            console.log('Expand/collapse button clicked');

            // Find the row element
            const row = button.closest('[role="row"]');
            if (!row) {
                return;
            }

            // Get aria-rowindex to identify the row
            const rowIndexStr = row.getAttribute('aria-rowindex');
            if (!rowIndexStr) {
                return;
            }

            const rowIndex = parseInt(rowIndexStr, 10);
            console.log('Row index:', rowIndex);

            // Look up snapshot ID from our mapping
            const snapshotId = rowIndexToIdRef.current.get(rowIndex / 2);

            if (snapshotId) {
                console.warn('Detected snapshot ID from row:', snapshotId, rowIndex);
                handleRowToggle(snapshotId);
            } else {
                console.warn('No snapshot ID found for row index:', rowIndex);
            }
        };

        // Use event delegation on the document to catch all button clicks
        document.addEventListener('click', handleExpandButtonClick, true);

        return () => {
            document.removeEventListener('click', handleExpandButtonClick, true);
        };
    }, [handleRowToggle]);

    // Build combined frame with summary + detail rows
    const combinedFrame = useMemo(() => {
        if (!summaryFrame) {
            return null;
        }

        const combined = new MutableDataFrame({
            refId: summaryFrame.refId,
            fields: [],
        });


        // Add all summary fields with data links for 'id' field
        summaryFrame.fields.forEach((field) => {
            const fieldConfig = {...field.config};

            // Add click handler to 'id' field
            if (field.name === 'id') {
                fieldConfig.links = [
                    {
                        title: 'Toggle Details',
                        url: '',
                        onClick: (event: any) => {
                            const snapshotId = String(event.origin.field.values.get(event.origin.rowIndex) ?? '');
                            panelEventBus.publish(
                                new TableRowClickEvent({
                                    id: snapshotId,
                                    rowIndex: event.origin.rowIndex,
                                })
                            );
                        },
                    },
                ];
            }

            combined.addField({
                name: field.name,
                type: field.type,
                config: fieldConfig,
                values: [],
            });
        });

        // Add detail-only fields (if any expanded rows have details)
        const allDetailFieldNames = new Set<string>();
        expandedRows.forEach((snapId) => {
            const details = detailsCache.get(snapId);
            if (details && details[0]) {
                details[0].fields.forEach((f) => {
                    if (!summaryFrame.fields.find((sf) => sf.name === f.name)) {
                        allDetailFieldNames.add(f.name);
                    }
                });
            }
        });

        allDetailFieldNames.forEach((fieldName) => {
            // Find first detail frame that has this field to get type/config
            let fieldType = FieldType.string;
            let fieldConfig = {};
            for (const snapId of expandedRows) {
                const details = detailsCache.get(snapId);
                if (details && details[0]) {
                    const field = details[0].fields.find((f) => f.name === fieldName);
                    if (field) {
                        fieldType = field.type;
                        fieldConfig = field.config;
                        break;
                    }
                }
            }
            combined.addField({
                name: fieldName,
                type: fieldType,
                config: fieldConfig,
                values: [],
            });
        });

        // Clear and rebuild row index mapping
        rowIndexToIdRef.current.clear();

        // Add summary rows
        const summaryRowCount = summaryFrame.length ?? summaryFrame.fields[0]?.values.length ?? 0;
        for (let rowIdx = 0; rowIdx < summaryRowCount; rowIdx++) {
            const snapId = summaryFrame.fields[idFieldIndex]?.values.get(rowIdx);

            const rowData: any = {};
            rowData['_rowType'] = 'summary';
            summaryFrame.fields.forEach((field) => {
                rowData[field.name] = field.values.get(rowIdx);
            });
            // Fill detail fields with null for summary rows
            allDetailFieldNames.forEach((fieldName) => {
                rowData[fieldName] = null;
            });

            // Map aria-rowindex (1-based) to snapshot ID
            if (snapId) {
                rowIndexToIdRef.current.set(rowIdx + 1, String(snapId));
            }

            combined.add(rowData);

            // If this row is expanded, add detail rows
            if (snapId && expandedRows.has(String(snapId))) {
                const details = detailsCache.get(String(snapId));
                if (details && details[0]) {
                    const detailFrame = details[0];
                    const detailRowCount = detailFrame.length ?? detailFrame.fields[0]?.values.length ?? 0;

                    for (let detailIdx = 0; detailIdx < detailRowCount; detailIdx++) {
                        const detailRowData: any = {};
                        detailRowData['_rowType'] = 'detail';

                        // Copy all summary fields to detail rows (for grouping)
                        summaryFrame.fields.forEach((field) => {
                            detailRowData[field.name] = field.values.get(rowIdx);
                        });

                        // Add/override with detail fields
                        detailFrame.fields.forEach((field) => {
                            detailRowData[field.name] = field.values.get(detailIdx);
                        });

                        combined.add(detailRowData);
                    }
                }
            }
        }

        return combined;
    }, [summaryFrame, expandedRows, detailsCache, idFieldIndex, panelEventBus]);

    // Apply groupToNestedTable transformation
    useEffect(() => {
        if (!combinedFrame) {
            setTransformedFrames([]);
            return;
        }

        console.log('Combined frame before transform:', combinedFrame);
        console.log('Combined frame fields:', combinedFrame.fields.map(f => f.name));
        console.log('Combined frame row count:', combinedFrame.length);

        // Build fields config with all summary fields as groupby
        const fieldsConfig: Record<string, any> = {};
        summaryFrame.fields.forEach((field) => {
            fieldsConfig[field.name] = {
                operation: 'groupby',
                aggregations: [],
            };
        });

        const transformer: DataTransformerConfig = {
            id: 'groupToNestedTable',
            options: {
                fields: fieldsConfig,
                showSubframeHeaders: true,
            },
        };
        const transformer2:DataTransformerConfig =
            {
                id: "organize",
                options: {
                    excludeByName: {
                        waitsByType: true
                    },
                    includeByName: {},
                    indexByName: {},
                    renameByName: {}
                }
            }
        try {
            const result = transformDataFrame([transformer, transformer2], [combinedFrame]);
            console.log('Transform result:', result);
            console.log('Transform result type:', typeof result);

            // transformDataFrame returns Observable<DataFrame[]>
            if (result && typeof result === 'object' && 'subscribe' in result) {
                const subscription = (result as Observable<DataFrame[]>).subscribe({
                    next: (frames) => {
                        console.log('Transformed frames:', frames);
                        console.log('Transformed frames count:', frames.length);
                        frames.forEach((frame, i) => {
                            console.log(`Frame ${i}:`, frame);
                            console.log(`Frame ${i} meta:`, frame.meta);
                        });
                        setTransformedFrames(frames);
                    },
                    error: (err) => {
                        console.error('Transformation Observable error:', err);
                        setTransformedFrames([combinedFrame]);
                    },
                });

                return () => subscription.unsubscribe();
            } else {
                console.warn('Transform returned non-Observable, using combined frame');
                setTransformedFrames([combinedFrame]);
            }
        } catch (err) {
            console.error('Transformation failed:', err);
            setTransformedFrames([combinedFrame]);
        }
        return;
    }, [combinedFrame, summaryFrame?.fields]);

    if (!summaryFrame) {
        return <div>No summary data</div>;
    }

    const panelData: PanelData = {
        series: Array.isArray(transformedFrames) ? transformedFrames : [],
        state: detailsLoading.size > 0 ? LoadingState.Loading : LoadingState.Done,
        timeRange: timeRange,
    };

    console.log('Rendering PanelData:', panelData);

    let fieldConfig1: FieldConfigSource = {
        defaults: {},

        overrides: [
            {
                matcher: {id: 'byName', options: 'Waits'},
                "properties": [
                    {
                        "id": "custom.cellOptions",
                        "value": {
                            "type": "markdown"
                        }
                    }
                ],
            },
            {
                matcher: {id: 'byName', options: 'waaaaa'},
                "properties": [
                    {
                        "id": "custom.cellOptions",
                        "value": {
                            "type": "markdown"
                        }
                    }
                ],
            },
        ],
    };
    return (
        <PanelContextProvider value={panelContext}>
            <div>
                <PanelRenderer
                    title="Nested Snapshots Table"
                    pluginId="table"
                    width={1200}
                    height={600}
                    data={panelData}
                    options={{
                        showHeader: true,
                        cellHeight: 'sm',
                        footer: {
                            show: false,
                            countRows: false,
                        },
                        frameIndex: 0,
                    }}
                    fieldConfig={fieldConfig1}
                />
            </div>
        </PanelContextProvider>
    );
}
