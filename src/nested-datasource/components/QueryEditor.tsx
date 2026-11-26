import React, {useState, useEffect} from 'react';
import {QueryEditorProps} from '@grafana/data';
import {Combobox, InlineField, InlineFieldRow, ComboboxOption, Input, MultiCombobox} from '@grafana/ui';
import {DataSource} from '../datasource';
import {MyDataSourceOptions, MyQuery} from '../types';

type Props = QueryEditorProps<DataSource, MyQuery, MyDataSourceOptions>;

export function QueryEditor({query, onChange, onRunQuery, datasource}: Props) {
    const [databaseOptions, setDatabaseOptions] = useState<ComboboxOption[]>([]);

    // Load database options on component mount
    useEffect(() => {
        datasource.getDatabaseOptions().then(setDatabaseOptions);
    }, [datasource]);


    const onDatabaseChange = (value: ComboboxOption<string | number> | null) => {
        const databaseValue = value?.value ? String(value.value) : undefined;

        onChange({...query, database: databaseValue});
    };
    const onQueryTypeChange = (value: ComboboxOption<string | number> | null) => {
        const databaseValue = value?.value ? String(value.value) : undefined;

        onChange({...query, queryType: databaseValue});
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
                            {label: "plan_analysis", value: "plan_analysis"},
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
                {(query.queryType === "snapshot") && (
                    <InlineField label="Snapshot ID">
                        <Input
                            value={query.snapshotID}
                            onChange={event => {
                                onChange({...query, snapshotID: event.currentTarget.value || ''})
                            }}
                            placeholder="Select snapshot"/>
                    </InlineField>
                )}

                {(query.queryType === "plan_analysis") && (
                    <>
                        <InlineField label="Databases (CSV)">
                            <Input
                                value={(query.databases || []).join(',')}
                                onChange={event => {
                                    const list = event.currentTarget.value
                                        .split(',')
                                        .map(s => s.trim())
                                        .filter(Boolean);
                                    onChange({...query, databases: list});
                                }}
                                placeholder="db1,db2"/>
                        </InlineField>
                        <InlineField label="Issue Types">
                            <MultiCombobox
                                width={"auto"}
                                minWidth={24}
                                onChange={option => onChange({...query, issueTags: option.map(value => value.value)})}
                                options={[
                                    {label: "Seek Plan Convert", value: "seek_plan_convert"},
                                    {label: "Cardinality Convert", value: "cardinality_convert"},
                                    {label: "Missing Index", value: "missing_index"},
                                    {label: "Implicit Conversion (Any)", value: "implicit_conversion"},
                                    {label: "Large Table Scan", value: "large_table_scan"},
                                ]}
                                value={query.issueTags}
                            />
                        </InlineField>
                    </>
                )}
                {(query.queryType === "metrics_series") && (
                    <InlineField label="Query Hash">
                        <Input
                            value={query.queryHash}
                            onChange={event => {
                                onChange({...query, queryHash: event.currentTarget.value || ''})
                            }}
                            placeholder=""/>
                    </InlineField>
                )}
                <InlineField label="Metrics">
                    <MultiCombobox
                        width={"auto"}
                        minWidth={20}
                        onChange={option => onChange({...query, metrics: option.map(value => value.value)})}
                        options={[
                            {label: "executionCount", value: "executionCount"},
                            {label: "avgElapsedTime", value: "avgElapsedTime"},
                            {label: "avgWorkerTime", value: "avgWorkerTime"},
                            {label: "avgRows", value: "avgRows"},
                            {label: "avgColumnstoreSegmentSkips", value: "avgColumnstoreSegmentSkips"},
                            {label: "avgReservedThreads", value: "avgReservedThreads"},
                            {label: "avgSpills", value: "avgSpills"},
                            {label: "avgGrantKb", value: "avgGrantKb"},
                            {label: "avgDop", value: "avgDop"},
                            {label: "avgLogicalReads", value: "avgLogicalReads"},
                            {label: "avgUsedThreads", value: "avgUsedThreads"},
                            {label: "avgPhysicalReads", value: "avgPhysicalReads"},
                            {label: "avgClrTime", value: "avgClrTime"},
                            {label: "avgUsedGrantKb", value: "avgUsedGrantKb"},
                            {label: "avgColumnstoreSegmentReads", value: "avgColumnstoreSegmentReads"},
                            {label: "avgLogicalWrites", value: "avgLogicalWrites"},
                        ]}
                        value={query.metrics}
                    />
                </InlineField>

            </InlineFieldRow>
            {/* Add other query fields as needed */}
        </>
    );
}
