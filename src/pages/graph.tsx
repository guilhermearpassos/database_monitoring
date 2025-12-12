import {
    GraphDrawStyle
} from "@grafana/schema";
import {GraphFieldConfig, PanelContext, PanelContextProvider, StackingMode, useTheme2} from "@grafana/ui";
import {
    applyFieldOverrides,
    DataFrame, dateTime,
    EventBus, FieldColorModeId, FieldConfigSource,
    LoadingState, PanelData,

} from "@grafana/data";
import {PanelRenderer} from "@grafana/runtime";
import React, {useMemo, useState} from "react";

// export const GRAPH_STYLES = ['lines', 'bars', 'points', 'stacked_lines', 'stacked_bars'] as const;
// export type GraphStyle = (typeof GRAPH_STYLES)[number];

interface Props {
    data: DataFrame[];
    loadingState: LoadingState;
    eventBus: EventBus;
}

export function MyGraph ({
                   data,
                   loadingState,
                   eventBus,
               }: Props) {
    const theme = useTheme2();

    const [fieldConfig, _] = useState<FieldConfigSource<GraphFieldConfig>>({
        defaults: {
            min: undefined,
            max: undefined,
            unit: 'short',
            color: {
                mode: FieldColorModeId.PaletteClassic,
            },
            custom: {
                pointSize: 5,
                drawStyle: GraphDrawStyle.Bars,
                stacking: {
                    mode: StackingMode.Normal,
                    group: 'A'
                },
                fillOpacity: 100,
                axisSoftMin: 0,
            },
        },
        overrides: [],
    });
    const panelContext: PanelContext = useMemo(() => ({
        eventBus: eventBus,
        onInstanceStateChange: () => {
        },
        canAddAnnotations: () => false,
        canEditAnnotations: () => false,
        eventsScope: "sqlsights-one",
        canDeleteAnnotations: () => false,
        // onToggleSeriesVisibility(label: string, mode: SeriesVisibilityChangeMode) {
        //     setFieldConfig(seriesVisibilityConfigFactory(label, mode, fieldConfig, data));
        // },
    }), [eventBus]);


    const dataWithConfig = useMemo(() => {
        return applyFieldOverrides({
            fieldConfig: fieldConfig,
            data: data,
            timeZone: 'browser',
            replaceVariables: (value) => value, // We don't need proper replace here as it is only used in getLinks and we use getFieldLinks
            theme,
        });
    }, [data, theme, fieldConfig]);

    const panelData: PanelData = {
        series: dataWithConfig, state: LoadingState.Loading,
        timeRange: {
            from: dateTime().subtract(1, 'hour'),
            to: dateTime(),
            raw: {from: 'now-1h', to: 'now'}
        }

    }
    return (

        <PanelContextProvider value={panelContext}>
            <PanelRenderer
                title="My Graph"
                pluginId="timeseries" // or "graph" for the old Graph panel
                width={800}
                height={400}
                data={panelData}
                fieldConfig={fieldConfig}

            />
        </PanelContextProvider>
    )
}
