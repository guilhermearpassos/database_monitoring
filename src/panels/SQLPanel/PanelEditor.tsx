import React from 'react';
import { StandardEditorProps } from '@grafana/data';
import { SQLPanelOptions } from './types';
import { Field, Input, Select, Switch } from '@grafana/ui';

interface Props extends StandardEditorProps<SQLPanelOptions> {}

export const PanelEditor: React.FC<Props> = ({ value: options, onChange }) => {
    const displayModeOptions = [
        { label: 'Table', value: 'table' },
        { label: 'Chart', value: 'chart' },
        { label: 'Metric', value: 'metric' },
    ];

    const chartTypeOptions = [
        { label: 'Line', value: 'line' },
        { label: 'Bar', value: 'bar' },
        { label: 'Pie', value: 'pie' },
    ];

    const updateOption = <T extends keyof SQLPanelOptions>(key: T, value: SQLPanelOptions[T]) => {
        onChange({
            ...options,
            [key]: value,
        });
    };

    return (
        <div>
            <Field label="Display Mode">
                <Select
                    value={options.displayMode}
                    options={displayModeOptions}
                    onChange={(option) => updateOption('displayMode', option.value as any)}
                />
            </Field>

            <Field label="Show Header">
                <Switch
                    value={options.showHeader}
                    onChange={(e) => updateOption('showHeader', e.currentTarget.checked)}
                />
            </Field>

            {options.displayMode === 'table' && (
                <>
                    <Field label="Show Row Numbers">
                        <Switch
                            value={options.showRowNumbers}
                            onChange={(e) => updateOption('showRowNumbers', e.currentTarget.checked)}
                        />
                    </Field>

                    <Field label="Alternate Row Colors">
                        <Switch
                            value={options.alternateRowColors}
                            onChange={(e) => updateOption('alternateRowColors', e.currentTarget.checked)}
                        />
                    </Field>
                </>
            )}

            {options.displayMode === 'chart' && (
                <Field label="Chart Type">
                    <Select
                        value={options.chartType}
                        options={chartTypeOptions}
                        onChange={(option) => updateOption('chartType', option.value as any)}
                    />
                </Field>
            )}

            {options.displayMode === 'metric' && (
                <>
                    <Field label="Unit">
                        <Input
                            value={options.unit || ''}
                            onChange={(e) => updateOption('unit', e.currentTarget.value)}
                            placeholder="e.g., ms, MB, %"
                        />
                    </Field>

                    <Field label="Decimals">
                        <Input
                            type="number"
                            value={options.decimals || 2}
                            onChange={(e) => updateOption('decimals', parseInt(e.currentTarget.value, 10))}
                        />
                    </Field>
                </>
            )}

            <Field label="Font Size">
                <Input
                    type="number"
                    value={options.fontSize}
                    onChange={(e) => updateOption('fontSize', parseInt(e.currentTarget.value, 10))}
                />
            </Field>

            <Field label="Max Rows">
                <Input
                    type="number"
                    value={options.maxRows}
                    onChange={(e) => updateOption('maxRows', parseInt(e.currentTarget.value, 10))}
                />
            </Field>
        </div>
    );
};
