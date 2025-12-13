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
import React, {useMemo, useState} from "react";

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
}

function setSeriesHiddenInConfig(
    cfg: FieldConfigSource<GraphFieldConfig>,
    label: string,
    mode: SeriesVisibilityChangeMode
): FieldConfigSource<GraphFieldConfig> {
    // This mirrors the dashboard behavior at a high level by expressing visibility as field overrides.
    // Time series panel understands `custom.hideFrom.viz`.
    const hidePropId = "custom.hideFrom";
    const hiddenValue = {viz: true, legend: false, tooltip: false};

    const overrides = cfg.overrides ?? [];

    const withMatcher = (name: string) => ({
        matcher: {id: "byName", options: name},
        properties: [{id: hidePropId, value: hiddenValue}],
    });

    // Helper to remove any existing override for a given label
    const removeOverrideFor = (name: string) =>
        overrides.filter((o) => !(o?.matcher?.id === "byName" && o?.matcher?.options === name));

    if (mode === SeriesVisibilityChangeMode.ToggleSelection) {
        const existing = overrides.find((o) => o?.matcher?.id === "byName" && o?.matcher?.options === label);
        const currentlyHidden = existing?.properties?.some((p) => p.id === hidePropId && p.value?.viz === true);

        // Toggle: if hidden -> remove override; if shown -> add override
        return {
            ...cfg,
            overrides: currentlyHidden ? removeOverrideFor(label) : [...removeOverrideFor(label), withMatcher(label)],
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
                            width = 800,
                            height = 400,
                        }: Props) {
    const theme = useTheme2();

    const [fieldConfig, setFieldConfig] = useState<FieldConfigSource<GraphFieldConfig>>({
        defaults: {
            min: undefined,
            max: undefined,
            unit: "short",
            color: {
                mode: FieldColorModeId.PaletteClassic,
            },
            custom: {
                pointSize: 5,
                drawStyle: GraphDrawStyle.Bars,
                stacking: {
                    mode: StackingMode.Normal,
                    group: "A",
                },
                fillOpacity: 100,
                axisSoftMin: 0,
            },
        },
        overrides: [],
    });

    const panelContext: PanelContext = useMemo(
        () => ({
            eventBus: eventBus,
            onInstanceStateChange: () => {
            },
            canAddAnnotations: () => false,
            canEditAnnotations: () => false,
            canDeleteAnnotations: () => false,
            eventsScope: "sqlsights-one",
            // Enables box-select / zoom to call back into your app
            onChangeTimeRange: (range: TimeRange) => {
                onTimeRangeChange?.(range);
            },

            // Enables legend click-to-hide behavior
            onToggleSeriesVisibility: (label: string, mode: SeriesVisibilityChangeMode) => {
                setFieldConfig((prev) => setSeriesHiddenInConfig(prev, label, mode));
            },
        }),
        [eventBus, onTimeRangeChange]
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

    const panelData: PanelData = {
        series: dataWithConfig,
        state: loadingState,
        timeRange,
        // ... existing code ...
    };

    let handleOnChangeTimeRange = (abs: AbsoluteTimeRange) => {
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
        })
    };
    return (
        <PanelContextProvider value={panelContext}>
            <PanelRenderer
                title="My Graph"
                pluginId="timeseries"
                width={width}
                height={height}
                data={panelData}
                fieldConfig={fieldConfig}
                onChangeTimeRange={handleOnChangeTimeRange}

            />
        </PanelContextProvider>
    );
}
