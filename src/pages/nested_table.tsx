import React, { useEffect, useMemo, useState } from 'react';
import { PanelRenderer } from '@grafana/runtime';
import {
    BusEventWithPayload,
    DataFrame,
    EventBusSrv, getDisplayProcessor,
    LoadingState,
    PanelData,
    TimeRange,
} from '@grafana/data';
import {PanelContext, PanelContextProvider, Table, useTheme2} from '@grafana/ui';

// Define a proper event class
class TableRowClickEvent extends BusEventWithPayload<{ id: string; rowIndex: number }> {
  static type = 'table-row-click'; // Must be static!
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
  const theme = useTheme2();
  const [expandedRows, setExpandedRows] = useState<Set<string>>(new Set());
  const [detailsCache, setDetailsCache] = useState<Map<string, PanelData>>(new Map());

  const panelEventBus = useMemo(() => new EventBusSrv(), []);

  const panelContext: PanelContext = useMemo(
    () => ({
      eventBus: panelEventBus,
      eventsScope: 'sqlsights-one',
      onInstanceStateChange: () => {},
      canAddAnnotations: () => false,
      canEditAnnotations: () => false,
      canDeleteAnnotations: () => false,
    }),
    [panelEventBus]
  );

  const summaryFrame = summaryData?.[0];

  const idFieldIndex = useMemo(() => {
    if (!summaryFrame) {
      return -1;
    }
    return summaryFrame.fields.findIndex((f) => f.name === 'id');
  }, [summaryFrame]);

  const handleRowToggle = (snapshotId: string) => {
    setExpandedRows((prev) => {
      const next = new Set(prev);

      if (next.has(snapshotId)) {
        next.delete(snapshotId);
        return next;
      }

      next.add(snapshotId);

      // Fetch details if not cached
      if (!detailsCache.has(snapshotId)) {
        setDetailsCache((prevCache) => {
          const nextCache = new Map(prevCache);
            getDetailsData(snapshotId).then((details) => {nextCache.set(snapshotId, details)});
          return nextCache;
        });
      }

      return next;
    });
  };

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
  }, [panelEventBus]);

  if (!summaryFrame) {
    return <div>No summary data</div>;
  }

    const displayProcessors = useMemo(() => {
        if (!summaryFrame) {
            return [];
        }
        return summaryFrame.fields.map((field) => getDisplayProcessor({ field, theme }));
    }, [summaryFrame, theme]);

  const colCount = summaryFrame.fields.length;
  const rowCount = summaryFrame.length ?? summaryFrame.fields[0]?.values.length ?? 0;
    Table
  return (
    <div>
      <PanelContextProvider value={panelContext}>
        <div style={{ border: '1px solid #444', borderRadius: 4, overflow: 'hidden' }}>
          <table style={{ width: '100%', borderCollapse: 'collapse' }}>
            <thead>
              <tr style={{ background: '#1f1f1f' }}>
                {summaryFrame.fields.map((f) => (
                  <th
                    key={f.name}
                    style={{
                      textAlign: 'left',
                      padding: '8px 10px',
                      borderBottom: '1px solid #444',
                      fontWeight: 600,
                    }}
                  >
                    {f.name}
                  </th>
                ))}
              </tr>
            </thead>

            <tbody>
              {Array.from({ length: rowCount }).map((_, rowIndex) => {
                const snapshotId =
                  idFieldIndex >= 0 ? String(summaryFrame.fields[idFieldIndex].values.get(rowIndex) ?? '') : '';
                const isExpanded = snapshotId ? expandedRows.has(snapshotId) : false;

                return (
                  <React.Fragment key={snapshotId || String(rowIndex)}>
                    <tr
                      onClick={() => snapshotId && handleRowToggle(snapshotId)}
                      style={{
                        cursor: snapshotId ? 'pointer' : 'default',
                        background: isExpanded ? '#151515' : 'transparent',
                      }}
                    >
                      {summaryFrame.fields.map((f, i) => {
                          const raw = f.values.get(rowIndex);
                          const disp = displayProcessors[i]?.(raw);
                          const text = disp?.text ?? String(raw ?? '');

                          return (<td
                              key={`${f.name}-${rowIndex}`}
                              style={{
                                  padding: '8px 10px',
                                  borderBottom: '1px solid #2b2b2b',
                                  verticalAlign: 'top',
                                  whiteSpace: 'nowrap',
                              }}
                          >
                              {text}
                          </td>)
                      })}
                    </tr>

                    {isExpanded && (
                      <tr>
                        <td
                          colSpan={colCount}
                          style={{
                            padding: 0,
                            borderBottom: '1px solid #2b2b2b',
                          }}
                        >
                          <div
                            style={{
                              marginLeft: 24,
                              marginTop: 10,
                              marginBottom: 14,
                              marginRight: 10,
                              border: '1px solid #444',
                              padding: 10,
                              borderRadius: 4,
                            }}
                          >
                            <div style={{ marginBottom: 8, fontWeight: 600 }}>Details for Snapshot: {snapshotId}</div>

                            {(() => {
                              const detailData = detailsCache.get(snapshotId);
                              if (!detailData) {
                                return <div>Loading…</div>;
                              }

                              return (
                                <PanelRenderer
                                  title="detailsnestedtable"
                                  pluginId="table"
                                  width={1100}
                                  height={300}
                                  data={detailData}
                                />
                              );
                            })()}
                          </div>
                        </td>
                      </tr>
                    )}
                  </React.Fragment>
                );
              })}
            </tbody>
          </table>
        </div>
      </PanelContextProvider>
    </div>
  );
}
