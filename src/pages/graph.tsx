import {GraphDrawStyle} from "@grafana/schema";
import {
    GraphFieldConfig,
    PanelContext,
    PanelContextProvider,
    SeriesVisibilityChangeMode,
    StackingMode,
    useTheme2,
} from "@grafana/ui";
import {
    applyFieldOverrides,
    DataFrame,
    EventBus,
    FieldColorModeId,
    FieldConfigSource,
    LoadingState,
    PanelData,
    TimeRange,
    AbsoluteTimeRange, dateTime,
} from "@grafana/data";
import {PanelRenderer} from "@grafana/runtime";
import React, {useCallback, useEffect, useMemo, useState} from "react";
import {getWaitEventColorNoDefault} from "../utils/wait_event_colors";

// export const GRAPH_STYLES = ['lines', 'bars', 'points', 'stacked_lines', 'stacked_bars'] as const;
// export type GraphStyle = (typeof GRAPH_STYLES)[number];

interface Props {
    data: DataFrame[];
    loadingState: LoadingState;
    eventBus: EventBus;

    timeRange: TimeRange;
    onTimeRangeChange: (range: TimeRange) => void;

    width?: number;
    height?: number;
    drawStyle?: GraphDrawStyle;
    stacked?: boolean;
    unit?: string;
}

function setSeriesHiddenInConfig(
    cfg: FieldConfigSource<GraphFieldConfig>,
    label: string,
    mode: SeriesVisibilityChangeMode,
    allLabels: string[]
): FieldConfigSource<GraphFieldConfig> {
    const hidePropId = "custom.hideFrom";
    const hiddenValue = {viz: true, legend: false, tooltip: false};

    const overrides = cfg.overrides ?? [];

    const withMatcher = (name: string) => ({
        matcher: {id: "byName", options: name},
        properties: [{id: hidePropId, value: hiddenValue}],
    });

    const isHidden = (name: string) =>
        overrides.some(
            (o) =>
                o?.matcher?.id === "byName" &&
                o?.matcher?.options === name &&
                o?.properties?.some((p) => p.id === hidePropId && p.value?.viz === true)
        );

    const removeOverrideFor = (name: string) =>
        overrides.filter((o) => !(o?.matcher?.id === "byName" && o?.matcher?.options === name));

    if (mode === SeriesVisibilityChangeMode.AppendToSelection) {
        const currentlyHidden = isHidden(label);
        return {
            ...cfg,
            overrides: currentlyHidden ? removeOverrideFor(label) : [...removeOverrideFor(label), withMatcher(label)],
        };
    }

    if (mode === SeriesVisibilityChangeMode.ToggleSelection) {
        const others = allLabels.filter((n) => n !== label);
        const allOthersHidden = others.length > 0 && others.every((n) => isHidden(n)) && !isHidden(label);
        if (allOthersHidden) {
            return { ...cfg, overrides: [] };
        }
        return {
            ...cfg,
            overrides: others.map(withMatcher),
        };
    }

    return cfg;
}

export function MyGraph({
                            data,
                            loadingState,
                            eventBus,
                            timeRange,
                            onTimeRangeChange,
                            height: customHeight,
                            drawStyle = GraphDrawStyle.Bars,
                            stacked = true,
                            unit = "short",
                        }: Props) {
    const theme = useTheme2();

    const containerRef = React.useRef<HTMLDivElement>(null);
    const [dimensions, setDimensions] = useState({ width: 800, height: customHeight ?? 400});

    // Update dimensions when container resizes
    useEffect(() => {
        if (!containerRef.current) {
            return;
        }

        const resizeObserver = new ResizeObserver((entries) => {
            for (const entry of entries) {
                const { width, height } = entry.contentRect;
                if (width > 0 && height > 0) {
                    setDimensions({ width, height });
                }
            }
        });

        resizeObserver.observe(containerRef.current);
        return () => resizeObserver.disconnect();
    }, []);
    // Store visibility overrides separately
    const [visibilityOverrides, setVisibilityOverrides] = useState<FieldConfigSource<GraphFieldConfig>['overrides']>([]);

    const allSeriesLabels = useMemo(() => {
        const names: string[] = [];
        for (const frame of data) {
            for (const f of frame.fields ?? []) {
                if (f.type !== 'time' && f.name) {
                    names.push(f.name);
                }
            }
        }
        return names;
    }, [data]);

    const seriesKey = allSeriesLabels.join('\0');
    useEffect(() => {
        setVisibilityOverrides([]);
    }, [seriesKey]);

    // Compute fieldConfig with colors based on data
    const fieldConfig = useMemo<FieldConfigSource<GraphFieldConfig>>(() => {
        // Build color overrides from data
        const colorOverrides = [];
        const missing: string[] = [];

        if (data[0]?.fields) {
            for (const field of data[0].fields) {
                const color = getWaitEventColorNoDefault(field.name);
                if (!color) {
                    missing.push(field.name);
                    continue;
                }
                colorOverrides.push({
                    matcher: { id: "byName", options: field.name },
                    properties: [
                        {
                            id: "color",
                            value: {
                                mode: "fixed",
                                fixedColor: `rgba(${color})`,
                            }
                        }
                    ]
                });
            }
        }

        if (missing.length > 0) {
            console.warn("Missing colors for:", missing);
        }

        return {
            defaults: {
                min: undefined,
                max: undefined,
                unit,
                color: {
                    mode: FieldColorModeId.PaletteClassic,
                },
                custom: {
                    pointSize: 5,
                    drawStyle,
                    stacking: {
                        mode: stacked ? StackingMode.Normal : StackingMode.None,
                        group: "A",
                    },
                    fillOpacity: drawStyle === GraphDrawStyle.Bars ? 100 : 18,
                    axisSoftMin: 0,
                },
            },
            overrides: [...colorOverrides, ...visibilityOverrides],
        };
    }, [data, visibilityOverrides, drawStyle, stacked, unit]);

    const panelContext: PanelContext = useMemo(
        () => ({
            eventBus: eventBus,
            onInstanceStateChange: () => {
            },
            canAddAnnotations: () => false,
            canEditAnnotations: () => false,
            canDeleteAnnotations: () => false,
            eventsScope: "sqlsights-one",
            onChangeTimeRange: (range: TimeRange) => {
                onTimeRangeChange?.(range);
            },
            onToggleSeriesVisibility: (label: string | string[] | null,
             mode: SeriesVisibilityChangeMode
             ) => {
      if (label == null) {
        return;
      }
      const labels = Array.isArray(label) ? label : [label];
                setVisibilityOverrides((overrides) => {
                    const nextConfig = labels.reduce<FieldConfigSource<GraphFieldConfig>>(
                        (nextCfg, seriesLabel) => setSeriesHiddenInConfig(nextCfg, seriesLabel, mode, allSeriesLabels),
                        {
                            defaults: {},
                            overrides,
                        }
                    );

                    return nextConfig.overrides ?? [];
                });

            },
        }),
        [eventBus, onTimeRangeChange, allSeriesLabels]
    );
//  useEffect(() => {
//    if (!onTimeRangeChange) {
//      return;
//    }

//    // When the user box-selects / zooms on the time series panel, it emits a zoom event.
//    const sub = eventBus.getStream(PanelEvents.zoom).subscribe((evt: any) => {
//      const nextRange = evt?.payload?.timeRange;
//      if (nextRange) {
//        onTimeRangeChange(nextRange);
//      }
//    });

//    return () => sub.unsubscribe();
//  }, [eventBus, onTimeRangeChange]);

    const dataWithConfig = useMemo(() => {
        return applyFieldOverrides({
            fieldConfig: fieldConfig,
            data: data,
            timeZone: "browser",
            replaceVariables: (value) => value, // We don't need proper replace here as it is only used in getLinks and we use getFieldLinks
            theme,
        });
    }, [data, theme, fieldConfig]);

    const panelData: PanelData = useMemo(
        () => ({
            series: dataWithConfig,
            state: loadingState,
            timeRange,
        }),
        [dataWithConfig, loadingState, timeRange]
    );

    const handleOnChangeTimeRange = useCallback(
        (abs: AbsoluteTimeRange) => {
            const from = dateTime(abs.from);
            const to = dateTime(abs.to);
            onTimeRangeChange({
                from,
                to,
                raw: {
                    // Keep raw as absolute epoch ms to avoid downstream parsing issues
                    from: from.toISOString(),
                    to: to.toISOString(),
                },
            });
        },
        [onTimeRangeChange]
    );
    return (
        <div ref={containerRef} style={{ width: '100%', height: customHeight ? `${customHeight}px` : '40vh' }}>
        <PanelContextProvider value={panelContext}>
            <PanelRenderer
                title="My Graph"
                pluginId="timeseries"
                width={dimensions.width}
                height={dimensions.height}
                data={panelData}
                fieldConfig={fieldConfig}
                onChangeTimeRange={handleOnChangeTimeRange}

            />
        </PanelContextProvider>
        </div>
    );
}
