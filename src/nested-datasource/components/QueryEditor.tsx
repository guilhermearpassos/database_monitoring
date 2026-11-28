import React, { useEffect, useState } from 'react';
import { css } from '@emotion/css';
import { GrafanaTheme2, QueryEditorProps } from '@grafana/data';
import {
  Combobox,
  ComboboxOption,
  InlineField,
  InlineFieldRow,
  InlineSwitch,
  Input,
  MultiCombobox,
  useStyles2,
} from '@grafana/ui';
import { DataSource } from '../datasource';
import { MyDataSourceOptions, MyQuery } from '../types';

type Props = QueryEditorProps<DataSource, MyQuery, MyDataSourceOptions>;

const QUERY_TYPE_OPTIONS: Array<ComboboxOption<string>> = [
  {
    label: 'Activity by Wait Event (Time Series)',
    value: 'chart',
    description: 'Stacked connections or wait time (ms) grouped by wait event over time',
  },
  {
    label: 'Connections & Blocking (Time Series)',
    value: 'blocking_timeseries',
    description: 'Active connections, blockers, waiters, and wait duration over time (ideal for alerts & trends)',
  },
  {
    label: 'Wait Events Summary (Aggregated)',
    value: 'wait_breakdown',
    description: 'Total connections, wait time (ms), and share % per wait event (ideal for Pie / Bar Gauge)',
  },
  {
    label: 'Snapshots List (Table)',
    value: 'snapshot-list',
    description: 'Snapshot summaries with deep links into Sqlsights',
  },
  {
    label: 'Snapshot Samples (Table)',
    value: 'snapshot',
    description: 'Active queries captured inside a specific snapshot ID',
  },
  {
    label: 'Bad Execution Plans (Table)',
    value: 'plan_analysis',
    description: 'Filterable table of plans with Seek Plan conversions, missing indexes, and lock counts',
  },
  {
    label: 'Plan Analysis KPIs (Stat / Summary)',
    value: 'plan_analysis_summary',
    description: 'Single-row KPI counts (Seek Plan converts, missing indexes, related locks) for Stat panels',
  },
  {
    label: 'Monitored Servers (Table)',
    value: 'servers',
    description: 'Fleet overview of all SQL Server instances active in the time range',
  },
  {
    label: 'Query Metrics Summary (Table)',
    value: 'metrics',
    description: 'Aggregated per-query execution counters',
  },
  {
    label: 'Query Metrics (Time Series)',
    value: 'metrics_series',
    description: 'Execution counters over time for a specific Query Hash',
  },
];

const METRIC_MODE_OPTIONS: Array<ComboboxOption<'connections' | 'wait_time_ms'>> = [
  { label: 'Active Connections', value: 'connections' },
  { label: 'Wait Time (ms)', value: 'wait_time_ms' },
];

const BLOCKING_SERIES_OPTIONS: Array<ComboboxOption<string>> = [
  { label: 'Connections', value: 'connections' },
  { label: 'Blockers', value: 'blockers' },
  { label: 'Waiters', value: 'waiters' },
  { label: 'Wait Duration (s)', value: 'waitDurationSec' },
];

const ISSUE_TAG_OPTIONS: Array<ComboboxOption<string>> = [
  { label: 'Implicit Conversion (Seek Plan)', value: 'seek_plan_convert' },
  { label: 'Missing Index', value: 'missing_index' },
  { label: 'Implicit Conversion (Cardinality)', value: 'cardinality_convert' },
  { label: 'Implicit Conversion (Any)', value: 'implicit_conversion' },
  { label: 'Large Table Scan', value: 'large_table_scan' },
];

const QUERY_METRIC_OPTIONS: Array<ComboboxOption<string>> = [
  { label: 'executionCount', value: 'executionCount' },
  { label: 'avgElapsedTime', value: 'avgElapsedTime' },
  { label: 'avgWorkerTime', value: 'avgWorkerTime' },
  { label: 'avgRows', value: 'avgRows' },
  { label: 'avgLogicalReads', value: 'avgLogicalReads' },
  { label: 'avgLogicalWrites', value: 'avgLogicalWrites' },
  { label: 'avgPhysicalReads', value: 'avgPhysicalReads' },
  { label: 'avgSpills', value: 'avgSpills' },
  { label: 'avgGrantKb', value: 'avgGrantKb' },
  { label: 'avgUsedGrantKb', value: 'avgUsedGrantKb' },
  { label: 'avgDop', value: 'avgDop' },
  { label: 'avgReservedThreads', value: 'avgReservedThreads' },
  { label: 'avgUsedThreads', value: 'avgUsedThreads' },
  { label: 'avgClrTime', value: 'avgClrTime' },
  { label: 'avgColumnstoreSegmentReads', value: 'avgColumnstoreSegmentReads' },
  { label: 'avgColumnstoreSegmentSkips', value: 'avgColumnstoreSegmentSkips' },
];

const getStyles = (theme: GrafanaTheme2) => ({
  hint: css`
    margin-top: ${theme.spacing(0.75)};
    font-size: ${theme.typography.bodySmall.fontSize};
    color: ${theme.colors.text.secondary};
  `,
});

export function QueryEditor({ query, onChange, onRunQuery, datasource }: Props) {
  const styles = useStyles2(getStyles);
  const [serverOptions, setServerOptions] = useState<Array<ComboboxOption<string>>>([]);
  const [serverDbOptions, setServerDbOptions] = useState<Array<ComboboxOption<string>>>([]);
  const [snapshotOptions, setSnapshotOptions] = useState<Array<ComboboxOption<string>>>([]);

  const queryType = query.queryType || 'chart';
  const needsServer = queryType !== 'servers';
  const isPlanQuery = queryType === 'plan_analysis' || queryType === 'plan_analysis_summary';

  useEffect(() => {
    let active = true;
    datasource.getDatabaseOptions().then((opts) => {
      if (active) {
        setServerOptions(opts);
      }
    });
    return () => {
      active = false;
    };
  }, [datasource]);

  useEffect(() => {
    let active = true;
    const server = query.database || '';
    if (!server || server.startsWith('$') || (!isPlanQuery && queryType !== 'snapshot')) {
      setServerDbOptions([]);
      return;
    }
    datasource.getServerDatabaseOptions(server).then((opts) => {
      if (active) {
        setServerDbOptions(opts);
      }
    });
    return () => {
      active = false;
    };
  }, [datasource, query.database, isPlanQuery, queryType]);

  useEffect(() => {
    let active = true;
    const server = query.database || '';
    if (!server || server.startsWith('$') || queryType !== 'snapshot') {
      setSnapshotOptions([]);
      return;
    }
    datasource.getSnapshotOptions(server).then((opts) => {
      if (active) {
        setSnapshotOptions(opts);
      }
    });
    return () => {
      active = false;
    };
  }, [datasource, query.database, queryType]);

  const updateAndRun = (patch: Partial<MyQuery>) => {
    onChange({ ...query, ...patch });
    onRunQuery();
  };

  return (
    <div>
      <InlineFieldRow>
        <InlineField label="Query Type" tooltip="Choose the Sqlsights dataset to query in Explore or Dashboards">
          <Combobox<string>
            width={34}
            options={QUERY_TYPE_OPTIONS}
            value={queryType}
            onChange={(opt) => updateAndRun({ queryType: opt?.value || 'chart' })}
            placeholder="Select query type"
          />
        </InlineField>

        {needsServer && (
          <InlineField
            label="SQL Server"
            tooltip="Target SQL Server instance alias (supports dashboard variables like $server)"
          >
            <Combobox<string>
              width={26}
              options={serverOptions}
              value={query.database || null}
              onChange={(opt) => updateAndRun({ database: opt?.value || '', databases: undefined })}
              placeholder="Select server or type $var..."
              createCustomValue
              isClearable
            />
          </InlineField>
        )}

        {(queryType === 'chart' || queryType === 'wait_breakdown') && (
          <InlineField label="Metric" tooltip="Measure active sessions or accumulated wait time (ms)">
            <Combobox<'connections' | 'wait_time_ms'>
              width={22}
              options={METRIC_MODE_OPTIONS}
              value={query.metric || 'connections'}
              onChange={(opt) => updateAndRun({ metric: opt?.value || 'connections' })}
            />
          </InlineField>
        )}

        {queryType === 'blocking_timeseries' && (
          <InlineField label="Series" tooltip="Leave empty to include all series">
            <MultiCombobox<string>
              width={32}
              options={BLOCKING_SERIES_OPTIONS}
              value={query.blockingSeries || []}
              onChange={(opts) => updateAndRun({ blockingSeries: opts.map((o) => o.value) })}
              placeholder="All series"
            />
          </InlineField>
        )}
      </InlineFieldRow>

      {(queryType === 'chart' || queryType === 'snapshot-list') && (
        <InlineFieldRow>
          {queryType === 'snapshot-list' && (
            <InlineField label="Only Blocking" tooltip="Return only snapshots with active blockers or waiters">
              <InlineSwitch
                value={Boolean(query.onlyBlocking)}
                onChange={(e) => updateAndRun({ onlyBlocking: e.currentTarget.checked })}
              />
            </InlineField>
          )}
          <InlineField label="Wait Event" tooltip="Optional wait event filter (e.g., LCK_M_X, CXPACKET, cpu)">
            <Input
              width={24}
              value={query.waitEvent || ''}
              onChange={(e) => onChange({ ...query, waitEvent: e.currentTarget.value })}
              onBlur={onRunQuery}
              onKeyDown={(e) => e.key === 'Enter' && onRunQuery()}
              placeholder="All wait events"
            />
          </InlineField>
        </InlineFieldRow>
      )}

      {queryType === 'snapshot' && (
        <InlineFieldRow>
          <InlineField label="Snapshot ID" tooltip="Pick a recent snapshot or enter a variable like $snapshot_id">
            <Combobox<string>
              width={36}
              options={snapshotOptions}
              value={query.snapshotID || null}
              onChange={(opt) => updateAndRun({ snapshotID: opt?.value || '' })}
              placeholder="Select snapshot or type $var..."
              createCustomValue
              isClearable
            />
          </InlineField>
          <InlineField label="Database(s)" tooltip="Optional filter by SQL Server database name(s)">
            <MultiCombobox<string>
              width={26}
              options={serverDbOptions}
              value={query.databases || []}
              onChange={(opts) => updateAndRun({ databases: opts.map((o) => o.value) })}
              placeholder="All databases"
              createCustomValue
            />
          </InlineField>
          <InlineField label="Only Blocking">
            <InlineSwitch
              value={Boolean(query.onlyBlocking)}
              onChange={(e) => updateAndRun({ onlyBlocking: e.currentTarget.checked })}
            />
          </InlineField>
        </InlineFieldRow>
      )}

      {isPlanQuery && (
        <InlineFieldRow>
          <InlineField
            label="Database(s)"
            tooltip="Filter by one or more databases on the instance (supports variables like $db)"
          >
            <MultiCombobox<string>
              width={28}
              options={serverDbOptions}
              value={query.databases || []}
              onChange={(opts) => updateAndRun({ databases: opts.map((o) => o.value) })}
              placeholder="All databases"
              createCustomValue
            />
          </InlineField>

          <InlineField label="Issue Types" tooltip="Filter by execution plan issue classification">
            <MultiCombobox<string>
              width={34}
              options={ISSUE_TAG_OPTIONS}
              value={query.issueTags || []}
              onChange={(opts) => updateAndRun({ issueTags: opts.map((o) => o.value) })}
              placeholder="All issue types"
            />
          </InlineField>

          <InlineField label="Only With Locks" tooltip="Include only plans involved in blocking locks">
            <InlineSwitch
              value={Boolean(query.onlyLocks)}
              onChange={(e) => updateAndRun({ onlyLocks: e.currentTarget.checked })}
            />
          </InlineField>

          <InlineField label="Min Impact %" tooltip="Minimum missing-index impact percentage">
            <Input
              type="number"
              width={12}
              min={0}
              max={100}
              value={query.minImpact ?? ''}
              onChange={(e) => {
                const raw = e.currentTarget.value;
                const parsed = raw === '' ? undefined : Number(raw);
                onChange({ ...query, minImpact: parsed !== undefined && Number.isFinite(parsed) ? parsed : undefined });
              }}
              onBlur={onRunQuery}
              onKeyDown={(e) => e.key === 'Enter' && onRunQuery()}
              placeholder="0"
            />
          </InlineField>

          <InlineField label="Search" tooltip="Match SQL text, conversion expression, or table name">
            <Input
              width={26}
              value={query.search || ''}
              onChange={(e) => onChange({ ...query, search: e.currentTarget.value })}
              onBlur={onRunQuery}
              onKeyDown={(e) => e.key === 'Enter' && onRunQuery()}
              placeholder="CONVERT_IMPLICIT, table..."
            />
          </InlineField>
        </InlineFieldRow>
      )}

      {queryType === 'metrics_series' && (
        <InlineFieldRow>
          <InlineField label="Query Hash" tooltip="SQL Server query hash (supports $queryHash)">
            <Input
              width={28}
              value={query.queryHash || ''}
              onChange={(e) => onChange({ ...query, queryHash: e.currentTarget.value })}
              onBlur={onRunQuery}
              onKeyDown={(e) => e.key === 'Enter' && onRunQuery()}
              placeholder="Base64 query hash"
            />
          </InlineField>
          <InlineField label="Metrics" tooltip="Leave empty to return all metrics">
            <MultiCombobox<string>
              width={32}
              options={QUERY_METRIC_OPTIONS}
              value={query.metrics || []}
              onChange={(opts) => updateAndRun({ metrics: opts.map((o) => o.value) })}
              placeholder="All metrics"
            />
          </InlineField>
        </InlineFieldRow>
      )}

      <div className={styles.hint}>
        Dashboard variable queries supported: <code>servers</code>, <code>databases($server)</code>,{' '}
        <code>snapshots($server)</code>, <code>issue_tags</code>.
      </div>
    </div>
  );
}
