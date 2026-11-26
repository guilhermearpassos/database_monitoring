import { buildQueryHistoryDataFrames } from './QueryHistoryChart';
import { QueryMetricsSeriesPoint } from './types';

describe('buildQueryHistoryDataFrames', () => {
  const samplePoints: QueryMetricsSeriesPoint[] = [
    {
      timestamp: '2026-11-26T12:00:00Z',
      execution_count: 10,
      total_elapsed_ms: 1500,
      avg_elapsed_ms: 150,
      total_worker_ms: 800,
      avg_worker_ms: 80,
      total_logical_reads: 5000,
      avg_logical_reads: 500,
      avg_physical_reads: 12,
      avg_logical_writes: 3,
      avg_rows: 20,
      total_spills: 2,
      avg_spills: 0.2,
      avg_grant_kb: 2048,
      avg_used_grant_kb: 512,
    },
    {
      timestamp: '2026-11-26T12:05:00Z',
      execution_count: 20,
      total_elapsed_ms: 2000,
      avg_elapsed_ms: 100,
      total_worker_ms: 1000,
      avg_worker_ms: 50,
      total_logical_reads: 8000,
      avg_logical_reads: 400,
      avg_physical_reads: 5,
      avg_logical_writes: 2,
      avg_rows: 15,
      total_spills: 0,
      avg_spills: 0,
      avg_grant_kb: 1024,
      avg_used_grant_kb: 768,
    },
  ];

  it('returns empty frames when there are no points', () => {
    const res = buildQueryHistoryDataFrames([], 'execution_count');
    expect(res.frames).toEqual([]);
  });

  it('builds duration & cpu frame with ms unit', () => {
    const res = buildQueryHistoryDataFrames(samplePoints, 'duration_cpu');
    expect(res.unit).toBe('ms');
    expect(res.frames).toHaveLength(1);
    expect(res.frames[0].fields.map((f) => f.name)).toEqual(['Time', 'Avg Elapsed (ms)', 'Avg CPU (ms)']);
    expect(res.frames[0].fields[1].values).toEqual([150, 100]);
  });

  it('builds memory grant vs used frame with deckbytes unit', () => {
    const res = buildQueryHistoryDataFrames(samplePoints, 'memory_grant');
    expect(res.unit).toBe('deckbytes');
    expect(res.frames[0].fields.map((f) => f.name)).toEqual(['Time', 'Avg Grant (KB)', 'Avg Used Grant (KB)']);
  });

  it('builds spills bar frame', () => {
    const res = buildQueryHistoryDataFrames(samplePoints, 'spills');
    expect(res.frames[0].fields[1].values).toEqual([2, 0]);
  });
});
