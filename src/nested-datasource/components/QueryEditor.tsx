
import React, { useState, useEffect } from 'react';
import { QueryEditorProps, SelectableValue } from '@grafana/data';
import { Select, InlineField, InlineFieldRow } from '@grafana/ui';
import { MyDataSource } from '../datasource';
import { MyDataSourceOptions, MyQuery } from '../types';

type Props = QueryEditorProps<MyDataSource, MyQuery, MyDataSourceOptions>;

export function QueryEditor({ query, onChange, onRunQuery, datasource }: Props) {
    const [databaseOptions, setDatabaseOptions] = useState<SelectableValue[]>([]);
    const [tableOptions, setTableOptions] = useState<SelectableValue[]>([]);
    const [columnOptions, setColumnOptions] = useState<SelectableValue[]>([]);

    // Load database options on component mount
    useEffect(() => {
        datasource.getDatabaseOptions().then(setDatabaseOptions);
    }, [datasource]);

    // Load table options when database changes
    useEffect(() => {
        if (query.database) {
            datasource.getTableOptions(query.database).then(setTableOptions);
        } else {
            setTableOptions([]);
        }
    }, [query.database, datasource]);

    // Load column options when table changes
    useEffect(() => {
        if (query.database && query.table) {
            datasource.getColumnOptions(query.database, query.table).then(setColumnOptions);
        } else {
            setColumnOptions([]);
        }
    }, [query.database, query.table, datasource]);

    const onDatabaseChange = (value: SelectableValue<string>) => {
        onChange({ ...query, database: value.value, table: undefined, column: undefined });
    };

    const onTableChange = (value: SelectableValue<string>) => {
        onChange({ ...query, table: value.value, column: undefined });
    };

    const onColumnChange = (value: SelectableValue<string>) => {
        onChange({ ...query, column: value.value });
    };

    return (
        <>
            <InlineFieldRow>
                <InlineField label="Database">
                    <Select
                        options={databaseOptions}
                        value={query.database}
                        onChange={onDatabaseChange}
                        placeholder="Select database"
                        width={20}
                    />
                </InlineField>
                <InlineField label="Table">
                    <Select
                        options={tableOptions}
                        value={query.table}
                        onChange={onTableChange}
                        placeholder="Select table"
                        width={20}
                        disabled={!query.database}
                    />
                </InlineField>
                <InlineField label="Column">
                    <Select
                        options={columnOptions}
                        value={query.column}
                        onChange={onColumnChange}
                        placeholder="Select column"
                        width={20}
                        disabled={!query.table}
                    />
                </InlineField>
            </InlineFieldRow>
            {/* Add other query fields as needed */}
        </>
    );
}