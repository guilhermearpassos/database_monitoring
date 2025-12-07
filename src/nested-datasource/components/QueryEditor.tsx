
import React, { useState, useEffect } from 'react';
import { QueryEditorProps, SelectableValue } from '@grafana/data';
import { Select, InlineField, InlineFieldRow } from '@grafana/ui';
import { DataSource } from '../datasource';
import { MyDataSourceOptions, MyQuery } from '../types';

type Props = QueryEditorProps<DataSource, MyQuery, MyDataSourceOptions>;

export function QueryEditor({ query, onChange, onRunQuery, datasource }: Props) {
    const [databaseOptions, setDatabaseOptions] = useState<SelectableValue[]>([]);

    // Load database options on component mount
    useEffect(() => {
        datasource.getDatabaseOptions().then(setDatabaseOptions);
    }, [datasource]);


    const onDatabaseChange = (value: SelectableValue<string>) => {
        onChange({ ...query, database: value.value });
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

            </InlineFieldRow>
            {/* Add other query fields as needed */}
        </>
    );
}