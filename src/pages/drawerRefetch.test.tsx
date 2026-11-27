import React from 'react';
import { MemoryRouter, Route, Routes } from 'react-router-dom';
import { fireEvent, render, screen, waitFor } from '@testing-library/react';
import { of } from 'rxjs';
import { getBackendSrv } from '@grafana/runtime';
import PageOne from './PageOne';
import PageTwo from './PageTwo';
import PageThree from './PageThree';
import PageFour from './PageFour';

jest.mock('@grafana/ui', () => {
  const actual = jest.requireActual('@grafana/ui');
  return {
    ...actual,
    TimeRangePicker: () => <div data-testid="time-range-picker" />,
    Drawer: ({
      title,
      onClose,
      children,
    }: {
      title?: React.ReactNode;
      onClose: () => void;
      children: React.ReactNode;
    }) => (
      <div role="dialog" aria-label={typeof title === 'string' ? title : 'Drawer'}>
        <button type="button" onClick={onClose}>
          Close Drawer
        </button>
        {children}
      </div>
    ),
  };
});

jest.mock('./graph', () => ({
  MyGraph: () => <div data-testid="mock-graph" />,
}));

jest.mock('@grafana/runtime', () => {
  const actual = jest.requireActual('@grafana/runtime');
  return {
    ...actual,
    getBackendSrv: jest.fn(),
  };
});

describe('Table auto-refetch regression on drawer close', () => {
  let fetchMock: jest.Mock;

  beforeEach(() => {
    fetchMock = jest.fn(({ url }: { url: string }) => {
      if (url.includes('/resources/datasource-options')) {
        return of({ data: [{ label: 'sql-prod', value: 'sql-prod' }] });
      }
      if (url.includes('/resources/snapshotSummaries')) {
        return of({
          data: [
            {
              id: 'snap-1',
              timestamp: '2026-11-27T10:00:00Z',
              connections: 5,
              blockers: 1,
              waiters: 2,
              wait_duration_ms: 1500,
              connections_by_wait_event: { LCK_M_X: 2, '': 3 },
              time_ms_by_wait_event: { LCK_M_X: 1500 },
            },
          ],
        });
      }
      if (url.includes('/resources/snapshotSamples')) {
        return of({
          data: [
            {
              snap_id: 'snap-1',
              sample_id: 'sample-1',
              session_id: '52',
              database: 'OrdersDB',
              status: 'suspended',
              wait_event: 'LCK_M_X',
              wait_time_ms: 1200,
              elapsed_ms: 2500,
              text: 'SELECT * FROM dbo.Orders WHERE Id = @p0',
              user: 'app_user',
              program_name: 'OrdersSvc',
              is_blocked: false,
              is_blocker: true,
              blocked_by: '',
              blocked_sessions: ['53'],
              query_hash: '0xabc',
            },
          ],
        });
      }
      if (url.includes('/resources/planAnalysis')) {
        return of({
          data: {
            plans: [
              {
                snap_id: 'snap-1',
                sample_id: 'sample-1',
                plan_handle: '0xplan1',
                query_hash: '0xabc',
                database: 'OrdersDB',
                text: 'SELECT * FROM dbo.Orders WHERE CustomerCode = @p0',
                issue_tags: ['seek_plan_convert'],
                issue_count: 1,
                occurrences: 4,
                related_lock_count: 2,
                max_impact: 82.5,
                last_snap_time: '2026-11-27T10:00:00Z',
                converts: [{ convert_issue: 'Seek Plan', expression: 'CONVERT_IMPLICIT(nvarchar(50), [CustomerCode])' }],
                missing_indexes: [],
              },
            ],
            available_databases: ['OrdersDB'],
          },
        });
      }
      if (url.includes('/resources/queryMetricsSeries')) {
        return of({ data: [] });
      }
      if (url.includes('/resources/queryMetrics')) {
        return of({
          data: {
            metrics: [
              {
                query_hash: '0xmetric1',
                database: 'OrdersDB',
                text: 'SELECT TOP 100 * FROM dbo.Orders ORDER BY CreatedAt DESC',
                execution_count: 120,
                total_elapsed_ms: 4200,
                avg_elapsed_ms: 35,
                last_elapsed_ms: 40,
                total_worker_ms: 2100,
                avg_worker_ms: 17.5,
                total_logical_reads: 50000,
                avg_logical_reads: 416.7,
                total_physical_reads: 10,
                avg_physical_reads: 0.08,
                total_logical_writes: 0,
                avg_logical_writes: 0,
                total_rows: 12000,
                total_spills: 0,
                avg_spills: 0,
                total_grant_kb: 122880,
                avg_grant_kb: 1024,
                avg_used_grant_kb: 512,
                avg_ideal_grant_kb: 512,
                memory_grant_waste_pct: 50,
                avg_rows: 100,
                avg_dop: 1,
                elapsed_share_pct: 45,
                cpu_share_pct: 40,
                reads_share_pct: 30,
                last_execution_time: '2026-11-27T10:00:00Z',
                issue_flags: [],
              },
              {
                query_hash: '0xmetric2',
                database: 'OrdersDB',
                text: 'SELECT * FROM dbo.BigArchiveTable',
                execution_count: 2,
                total_elapsed_ms: 2000,
                avg_elapsed_ms: 1000,
                last_elapsed_ms: 1000,
                total_worker_ms: 1000,
                avg_worker_ms: 500,
                total_logical_reads: 16000,
                avg_logical_reads: 8000,
                total_physical_reads: 4,
                avg_physical_reads: 2,
                total_logical_writes: 0,
                avg_logical_writes: 0,
                total_rows: 20000,
                total_spills: 4,
                avg_spills: 2,
                total_grant_kb: 20480,
                avg_grant_kb: 10240,
                avg_used_grant_kb: 2048,
                avg_ideal_grant_kb: 2048,
                memory_grant_waste_pct: 80,
                avg_rows: 10000,
                avg_dop: 4,
                elapsed_share_pct: 20,
                cpu_share_pct: 20,
                reads_share_pct: 10,
                last_execution_time: '2026-11-27T10:00:00Z',
                issue_flags: ['spills'],
              },
            ],
            available_databases: ['OrdersDB'],
          },
        });
      }
      if (url.includes('/resources/deadlocks')) {
        return of({
          data: {
            deadlocks: [
              {
                id: 'dl-1',
                server: 'sql-prod',
                timestamp: '2026-11-27T10:18:42Z',
                xml_hash: 'hash1',
                fingerprint: 'fp-rev',
                pattern_occurrences: 2,
                victim_spid: '64',
                databases: ['OrdersDB'],
                max_wait_ms: 3412,
                processes: [
                  {
                    id: 'p1',
                    is_victim: true,
                    spid: '64',
                    ecid: '0',
                    status: 'suspended',
                    wait_time_ms: 3412,
                    wait_resource: 'KEY: 6:1',
                    lock_mode: 'X',
                    log_used: 2416,
                    priority: '0',
                    tran_count: '2',
                    transaction_name: 'user_transaction',
                    isolation_level: 'read committed (2)',
                    client_app: 'OrdersAPI',
                    host_name: 'aks-orders',
                    login_name: 'svc_orders',
                    current_db: '6',
                    current_db_name: 'OrdersDB',
                    frames: [{ proc_name: 'dbo.usp_Reserve', line: '42', sql_handle: '0x01', sql: 'UPDATE dbo.Inventory SET Qty = Qty - 1;' }],
                    input_buf: 'EXEC dbo.usp_SubmitOrder;',
                    active_sql: 'UPDATE dbo.Inventory SET Qty = Qty - 1;',
                  },
                  {
                    id: 'p2',
                    is_victim: false,
                    spid: '71',
                    ecid: '0',
                    status: 'suspended',
                    wait_time_ms: 2980,
                    wait_resource: 'KEY: 6:2',
                    lock_mode: 'U',
                    log_used: 18640,
                    priority: '0',
                    tran_count: '2',
                    transaction_name: 'user_transaction',
                    isolation_level: 'read committed (2)',
                    client_app: 'InventoryWorker',
                    host_name: 'aks-worker',
                    login_name: 'svc_inv',
                    current_db: '6',
                    current_db_name: 'OrdersDB',
                    frames: [],
                    input_buf: 'UPDATE dbo.Orders SET Status = 1;',
                    active_sql: 'UPDATE dbo.Orders SET Status = 1;',
                  },
                ],
                resources: [
                  {
                    id: 'l1',
                    type: 'keylock',
                    dbid: '6',
                    database: 'OrdersDB',
                    object_name: 'OrdersDB.dbo.Inventory',
                    short_object: 'dbo.Inventory',
                    index_name: 'PK_Inventory',
                    hobt_id: '100',
                    file_id: '',
                    page_id: '',
                    slot: '',
                    mode: 'X',
                    owners: [{ process_id: 'p2', mode: 'X', request_type: '' }],
                    waiters: [{ process_id: 'p1', mode: 'X', request_type: 'wait' }],
                  },
                ],
                diagnosis: {
                  pattern: 'reverse_order',
                  pattern_label: 'Reverse-Order Lock Cycle',
                  severity: 'error',
                  summary: 'SPID 64 and 71 locked dbo.Orders and dbo.Inventory in opposite order.',
                  recommendations: ['Standardize lock acquisition order.'],
                  steps: [{ kind: 'victim', text: 'SPID 64 rolled back.' }],
                },
                xml_report: '<deadlock><victim-list><victimProcess id="p1"/></victim-list></deadlock>',
              },
              {
                id: 'dl-2',
                server: 'sql-prod',
                timestamp: '2026-11-27T10:12:00Z',
                xml_hash: 'hash2',
                fingerprint: 'fp-rev',
                pattern_occurrences: 2,
                victim_spid: '64',
                databases: ['OrdersDB'],
                max_wait_ms: 2100,
                processes: [],
                resources: [],
                diagnosis: {
                  pattern: 'reverse_order',
                  pattern_label: 'Reverse-Order Lock Cycle',
                  severity: 'error',
                  summary: 'Duplicate occurrence.',
                  recommendations: [],
                  steps: [],
                },
                xml_report: '<deadlock/>',
              },
            ],
            available_databases: ['OrdersDB'],
          },
        });
      }
      if (url.includes('/resources/getQueryDetails')) {
        return of({
          data: {
            query_sample: {
              snap_id: 'snap-1',
              sample_id: 'sample-1',
              sid: '52',
              database: 'OrdersDB',
              status: 'suspended',
              execution_time: '2.5s',
              wait_event: 'LCK_M_X',
              query: 'SELECT * FROM dbo.Orders WHERE Id = @p0',
              user: 'app_user',
              query_hash: '0xabc',
            },
          },
        });
      }
      return of({ data: [] });
    });

    (getBackendSrv as jest.Mock).mockReturnValue({ fetch: fetchMock });
  });

  const countFetches = (needle: string) =>
    fetchMock.mock.calls.filter(([req]) => typeof req?.url === 'string' && req.url.includes(needle)).length;

  test('Snapshots page does not re-fetch summaries or samples when opening/closing details drawer or sorting', async () => {
    render(
      <MemoryRouter initialEntries={['/snapshots?server=sql-prod']}>
        <Routes>
          <Route path="/snapshots" element={<PageOne />} />
        </Routes>
      </MemoryRouter>
    );

    await waitFor(() => expect(screen.getByText('1.50 s')).toBeInTheDocument());
    expect(countFetches('/resources/snapshotSummaries')).toBe(1);

    // Expand snapshot row
    fireEvent.click(screen.getByText('5').closest('tr')!);
    await waitFor(() => expect(screen.getByText('SELECT * FROM dbo.Orders WHERE Id = @p0')).toBeInTheDocument());
    expect(countFetches('/resources/snapshotSamples')).toBe(1);

    // Open details drawer
    fireEvent.click(screen.getByText('SELECT * FROM dbo.Orders WHERE Id = @p0').closest('tr')!);
    await waitFor(() => expect(screen.getByRole('dialog')).toBeInTheDocument());

    // Close details drawer
    fireEvent.click(screen.getByRole('button', { name: 'Close Drawer' }));
    await waitFor(() => expect(screen.queryByRole('dialog')).not.toBeInTheDocument());

    // Sort table
    fireEvent.click(screen.getByRole('button', { name: /Connections/i }));

    expect(countFetches('/resources/snapshotSummaries')).toBe(1);
    expect(countFetches('/resources/snapshotSamples')).toBe(1);
  });

  test('Plan Analysis page does not re-fetch plans when opening/closing details drawer, sorting, or searching', async () => {
    render(
      <MemoryRouter initialEntries={['/plan_analysis?server=sql-prod']}>
        <Routes>
          <Route path="/plan_analysis" element={<PageTwo />} />
        </Routes>
      </MemoryRouter>
    );

    await waitFor(() =>
      expect(screen.getByText('SELECT * FROM dbo.Orders WHERE CustomerCode = @p0')).toBeInTheDocument()
    );
    expect(countFetches('/resources/planAnalysis')).toBe(1);

    // Open details drawer
    fireEvent.click(screen.getByText('SELECT * FROM dbo.Orders WHERE CustomerCode = @p0').closest('tr')!);
    await waitFor(() => expect(screen.getByRole('dialog')).toBeInTheDocument());
    expect(countFetches('/resources/planAnalysis')).toBe(1);

    // Close details drawer
    fireEvent.click(screen.getByRole('button', { name: 'Close Drawer' }));
    await waitFor(() => expect(screen.queryByRole('dialog')).not.toBeInTheDocument());
    expect(countFetches('/resources/planAnalysis')).toBe(1);

    // Client-side sort and search should not re-fetch
    fireEvent.click(screen.getByRole('button', { name: /Occurrences/i }));
    fireEvent.change(screen.getByPlaceholderText(/Filter by SQL text/i), { target: { value: 'Orders' } });

    expect(countFetches('/resources/planAnalysis')).toBe(1);
    expect(screen.getByText('SELECT * FROM dbo.Orders WHERE CustomerCode = @p0')).toBeInTheDocument();
  });

  test('Query Metrics page does not re-fetch metrics when opening/closing details drawer, sorting, or searching', async () => {
    render(
      <MemoryRouter initialEntries={['/query_metrics?server=sql-prod']}>
        <Routes>
          <Route path="/query_metrics" element={<PageThree />} />
        </Routes>
      </MemoryRouter>
    );

    await waitFor(() =>
      expect(screen.getByText('SELECT TOP 100 * FROM dbo.Orders ORDER BY CreatedAt DESC')).toBeInTheDocument()
    );
    expect(countFetches('/resources/queryMetrics?')).toBe(1);

    // Open details drawer
    fireEvent.click(screen.getByText('SELECT TOP 100 * FROM dbo.Orders ORDER BY CreatedAt DESC').closest('tr')!);
    await waitFor(() => expect(screen.getByRole('dialog')).toBeInTheDocument());
    expect(countFetches('/resources/queryMetrics?')).toBe(1);

    // Close details drawer
    fireEvent.click(screen.getByRole('button', { name: 'Close Drawer' }));
    await waitFor(() => expect(screen.queryByRole('dialog')).not.toBeInTheDocument());
    expect(countFetches('/resources/queryMetrics?')).toBe(1);

    // Client-side sort and search should not re-fetch
    fireEvent.click(screen.getByRole('button', { name: /Execs/i }));
    fireEvent.change(screen.getByPlaceholderText(/Filter by SQL text/i), { target: { value: 'Orders' } });

    expect(countFetches('/resources/queryMetrics?')).toBe(1);
    expect(screen.getByText('SELECT TOP 100 * FROM dbo.Orders ORDER BY CreatedAt DESC')).toBeInTheDocument();
  });

  test('Query Metrics page toggles between Totals and Avg / Exec and sorts by average reads without re-fetching', async () => {
    render(
      <MemoryRouter initialEntries={['/query_metrics?server=sql-prod']}>
        <Routes>
          <Route path="/query_metrics" element={<PageThree />} />
        </Routes>
      </MemoryRouter>
    );

    await waitFor(() =>
      expect(screen.getByText('SELECT TOP 100 * FROM dbo.Orders ORDER BY CreatedAt DESC')).toBeInTheDocument()
    );

    const getRowOrder = () =>
      screen
        .getAllByRole('row')
        .slice(1)
        .map((row) => row.textContent || '');

    // Default Totals mode sorted by total_elapsed_ms desc: Orders (4200ms) before BigArchiveTable (2000ms)
    expect(getRowOrder()[0]).toContain('dbo.Orders');
    expect(getRowOrder()[1]).toContain('dbo.BigArchiveTable');

    // Sort by Logical Reads (total_logical_reads desc): Orders (50K) before BigArchiveTable (16K)
    fireEvent.click(screen.getByRole('button', { name: /Logical Reads/i }));
    expect(getRowOrder()[0]).toContain('dbo.Orders');
    expect(getRowOrder()[1]).toContain('dbo.BigArchiveTable');

    // Switch to Avg / Exec: active sort maps to avg_logical_reads desc, so BigArchiveTable (8,000 avg) jumps first
    fireEvent.click(screen.getByLabelText('Avg / Exec'));
    expect(screen.getByRole('button', { name: /Avg Reads/i })).toBeInTheDocument();
    expect(getRowOrder()[0]).toContain('dbo.BigArchiveTable');
    expect(getRowOrder()[1]).toContain('dbo.Orders');
    expect(getRowOrder()[0]).toContain('total 16.0K');

    // Switch back to Totals: active sort maps back to total_logical_reads desc, Orders jumps back first
    fireEvent.click(screen.getByLabelText('Totals'));
    expect(screen.getByRole('button', { name: /Logical Reads/i })).toBeInTheDocument();
    expect(getRowOrder()[0]).toContain('dbo.Orders');
    expect(getRowOrder()[1]).toContain('dbo.BigArchiveTable');

    // None of the mode/sort changes should trigger another backend fetch
    expect(countFetches('/resources/queryMetrics?')).toBe(1);
  });

  test('Deadlocks page renders events, groups by pattern, and opens/closes drawer without re-fetching', async () => {
    render(
      <MemoryRouter initialEntries={['/deadlocks?server=sql-prod']}>
        <Routes>
          <Route path="/deadlocks" element={<PageFour />} />
        </Routes>
      </MemoryRouter>
    );

    await waitFor(() => expect(screen.getByText('UPDATE dbo.Inventory SET Qty = Qty - 1;')).toBeInTheDocument());
    expect(countFetches('/resources/deadlocks')).toBe(1);

    // Events mode shows both deadlock rows
    expect(screen.getAllByRole('row')).toHaveLength(3); // 1 header + 2 rows

    // Switch to Grouped Patterns -> deduplicates by fingerprint into 1 row
    fireEvent.click(screen.getByLabelText('Grouped Patterns'));
    expect(screen.getAllByRole('row')).toHaveLength(2);
    expect(countFetches('/resources/deadlocks')).toBe(1);

    // Open details drawer
    fireEvent.click(screen.getByText('UPDATE dbo.Inventory SET Qty = Qty - 1;').closest('tr')!);
    await waitFor(() => expect(screen.getByRole('dialog')).toBeInTheDocument());
    expect(screen.getByText(/Automated Diagnosis: Reverse-Order Lock Cycle/)).toBeInTheDocument();
    expect(screen.getByTestId('deadlock-node-p1')).toBeInTheDocument();

    // Close details drawer -> must NOT re-fetch /resources/deadlocks
    fireEvent.click(screen.getByRole('button', { name: 'Close Drawer' }));
    await waitFor(() => expect(screen.queryByRole('dialog')).not.toBeInTheDocument());
    expect(countFetches('/resources/deadlocks')).toBe(1);
  });
});
