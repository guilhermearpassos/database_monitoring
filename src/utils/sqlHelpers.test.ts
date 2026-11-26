import { buildMissingIndexDDL, formatDurationMs } from './sqlHelpers';
import { buildTempoExploreHref } from '../context/AppSettingsContext';
import {
  applyTimeRangeToSearchParams,
  buildPreservedNavSearch,
  parseTimeRangeFromSearchParams,
} from './urlState';

describe('UI utilities', () => {
  test('buildMissingIndexDDL omits INCLUDE clause when empty', () => {
    const ddl = buildMissingIndexDDL({
      database: '[orders_db]',
      schema: '[dbo]',
      table: '[orders]',
      impact: 78.2,
      equality_columns: ['[customer_id]'],
      inequality_columns: ['[created_at]'],
      include_columns: [],
    });
    expect(ddl).toBe(
      'CREATE NONCLUSTERED INDEX [IX_orders_customer_id_created_at] ON [orders_db].[dbo].[orders] ([customer_id], [created_at]);'
    );
  });

  test('buildMissingIndexDDL includes INCLUDE clause when present', () => {
    const ddl = buildMissingIndexDDL({
      database: '[orders_db]',
      schema: '[dbo]',
      table: '[orders]',
      impact: 90.0,
      equality_columns: ['[status]'],
      include_columns: ['[total]', '[user_id]'],
    });
    expect(ddl).toContain('INCLUDE ([total], [user_id]);');
  });

  test('formatDurationMs formats milliseconds, seconds, and minutes', () => {
    expect(formatDurationMs(45)).toBe('45 ms');
    expect(formatDurationMs(2350)).toBe('2.35 s');
    expect(formatDurationMs(125000)).toBe('2m 5s');
  });

  test('buildTempoExploreHref builds a Grafana Explore traceql URL', () => {
    const href = buildTempoExploreHref('tempo-main', '0123456789abcdef0123456789abcdef');
    expect(href).toContain('/explore?');
    expect(href).toContain('tempo-main');
    expect(href).toContain('0123456789abcdef0123456789abcdef');
  });

  test('URL time range and preserved navigation helpers round-trip cleanly', () => {
    const params = new URLSearchParams('server=sql01&from=now-6h&to=now&db=orders_db');
    const tr = parseTimeRangeFromSearchParams(params);
    expect(tr.raw.from).toBe('now-6h');
    expect(tr.raw.to).toBe('now');

    const out = new URLSearchParams();
    applyTimeRangeToSearchParams(out, tr);
    expect(out.get('from')).toBe('now-6h');
    expect(out.get('to')).toBe('now');

    const nav = buildPreservedNavSearch(params, { snapId: 'snap-42' });
    expect(nav).toContain('server=sql01');
    expect(nav).toContain('from=now-6h');
    expect(nav).toContain('snapId=snap-42');
    expect(nav).not.toContain('db=orders_db');
  });
});
