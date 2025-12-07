import React from 'react';
import { PanelProps } from '@grafana/data';
import { SQLPanelOptions } from './types';
import { Table } from '@grafana/ui';

interface Props extends PanelProps<SQLPanelOptions> {}

export const SQLPanel: React.FC<Props> = ({ options, data, width, height }) => {
    if (!data || data.series.length === 0) {
        return (
            <div style={{ width, height, display: 'flex', alignItems: 'center', justifyContent: 'center' }}>
                <div>No data available</div>
            </div>
        );
    }

    const series = data.series[0];

    if (options.displayMode === 'table') {
        return (
            <div style={{ width, height }}>
                <Table
                    data={series}
                    width={width}
                    height={height}
                    showTypeIcons={true}
                    resizable={true}
                />
            </div>
        );
    }

    if (options.displayMode === 'metric') {
        const value = series.fields[0]?.values.get(0);
        return (
            <div
                style={{
                    width,
                    height,
                    display: 'flex',
                    alignItems: 'center',
                    justifyContent: 'center',
                    fontSize: `${options.fontSize * 2}px`,
                    fontWeight: 'bold'
                }}
            >
                {value} {options.unit}
            </div>
        );
    }

    return (
        <div style={{ width, height }}>
            <div>Chart visualization coming soon...</div>
        </div>
    );
};
