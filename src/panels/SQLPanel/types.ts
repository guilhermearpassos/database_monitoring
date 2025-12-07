
export interface SQLPanelOptions {
    showHeader: boolean;
    title?: string;
    refreshInterval: number;
    maxRows: number;
    displayMode: 'table' | 'chart' | 'metric';
    showRowNumbers: boolean;
    fontSize: number;
    headerColor?: string;
    alternateRowColors: boolean;
    chartType?: 'line' | 'bar' | 'pie';
    xAxisColumn?: string;
    yAxisColumn?: string;
    metricColumn?: string;
    unit?: string;
    decimals?: number;
    enableExport: boolean;
    enableFiltering: boolean;
}

export const defaultPanelOptions: SQLPanelOptions = {
    showHeader: true,
    refreshInterval: 30,
    maxRows: 100,
    displayMode: 'table',
    showRowNumbers: true,
    fontSize: 12,
    alternateRowColors: true,
    chartType: 'line',
    decimals: 2,
    enableExport: true,
    enableFiltering: true,
};
