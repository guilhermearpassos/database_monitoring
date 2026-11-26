import { DataSourceInstanceSettings } from '@grafana/data';
import { DataSource, parseVariableQuery } from './datasource';
import { MyDataSourceOptions } from './types';

describe('Sqlsights Nested DataSource', () => {
  test('parseVariableQuery parses simple and parameterized variable queries', () => {
    expect(parseVariableQuery('servers')).toEqual({ kind: 'servers', arg: undefined });
    expect(parseVariableQuery('databases(sql-prod-01)')).toEqual({ kind: 'databases', arg: 'sql-prod-01' });
    expect(parseVariableQuery('snapshots( $server )')).toEqual({ kind: 'snapshots', arg: '$server' });
    expect(parseVariableQuery('issue_tags')).toEqual({ kind: 'issue_tags', arg: undefined });
  });

  test('filterQuery allows servers without database and snapshot with snapshotID', () => {
    const ds = new DataSource({
      id: 1,
      uid: 'sqlsights-ds',
      name: 'Sqlsights',
      type: 'guilhermearpassos-sqlsights-datasource',
      url: '/api/datasources/proxy/1',
      jsonData: {},
    } as unknown as DataSourceInstanceSettings<MyDataSourceOptions>);

    expect(ds.filterQuery({ refId: 'A', queryType: 'servers' })).toBe(true);
    expect(ds.filterQuery({ refId: 'A', queryType: 'snapshot', snapshotID: 'snap-1' })).toBe(true);
    expect(ds.filterQuery({ refId: 'A', queryType: 'snapshot' })).toBe(false);
    expect(ds.filterQuery({ refId: 'A', queryType: 'chart', database: 'sql01' })).toBe(true);
    expect(ds.filterQuery({ refId: 'A', queryType: 'chart', database: '' })).toBe(false);
    expect(ds.filterQuery({ refId: 'A', queryType: 'servers', hide: true })).toBe(false);
  });
});
