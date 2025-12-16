
import React, { useState, useEffect } from 'react';
import { QueryEditorProps } from '@grafana/data';
import {Combobox, InlineField, InlineFieldRow, ComboboxOption, Input} from '@grafana/ui';
import { DataSource } from '../datasource';
import { MyDataSourceOptions, MyQuery } from '../types';

type Props = QueryEditorProps<DataSource, MyQuery, MyDataSourceOptions>;

export function QueryEditor({ query, onChange, onRunQuery, datasource }: Props) {
    const [databaseOptions, setDatabaseOptions] = useState<ComboboxOption[]>([]);

    // Load database options on component mount
    useEffect(() => {
        datasource.getDatabaseOptions().then(setDatabaseOptions);
    }, [datasource]);


    const onDatabaseChange = (value: ComboboxOption<string|number> | null) => {
        const databaseValue = value?.value ? String(value.value) : undefined;

        onChange({ ...query, database: databaseValue });
    };
    const onQueryTypeChange = (value: ComboboxOption<string|number> | null) => {
        const databaseValue = value?.value ? String(value.value) : undefined;

        onChange({ ...query, queryType: databaseValue });
    };


    return (
        <>
            <InlineFieldRow>
                <InlineField label="Query type">

                    <Combobox
                        width={"auto"}
                        options={[
                            {label: "chart", value: "chart"},
                            {label: "snapshot-list", value: "snapshot-list"},
                            {label: "snapshot", value: "snapshot"},
                            {label: "metrics", value: "metrics"},
                            {label: "metrics_series", value: "metrics_series"}
                        ]}
                        value={query.queryType}
                        onChange={onQueryTypeChange}
                        placeholder="Select query type"
                        minWidth={20}
                        isClearable={true}/>
                </InlineField>
                <InlineField label="Database">
                    <Combobox
                        width={"auto"}
                        options={databaseOptions}
                        value={query.database}
                        onChange={onDatabaseChange}
                        placeholder="Select database"
                        minWidth={20}
                        isClearable={true}/>
                </InlineField>
                {(query.queryType==="snapshot") && (
                    <InlineField label="Snapshot ID">
                        <Input
                            value={query.snapshotID}
                            onChange={event => {onChange({...query, snapshotID: event.currentTarget.value || ''})}}
                            placeholder="Select snapshot"/>
                    </InlineField>
                )}

                {(query.queryType==="metrics_series") && (
                    <InlineField label="Query Hash">
                        <Input
                            value={query.queryHash}
                            onChange={event => {onChange({...query, queryHash: event.currentTarget.value || ''})}}
                            placeholder=""/>
                    </InlineField>
                )}

            </InlineFieldRow>
            {/* Add other query fields as needed */}
        </>
    );
}
