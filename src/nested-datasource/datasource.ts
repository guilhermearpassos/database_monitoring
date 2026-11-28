import { FetchResponse, getBackendSrv, getTemplateSrv, isFetchError } from '@grafana/runtime';
import {
  CoreApp,
  DataQueryRequest,
  DataQueryResponse,
  DataSourceApi,
  DataSourceInstanceSettings,
  MetricFindValue,
  ScopedVars,
} from '@grafana/data';
import { lastValueFrom, Observable } from 'rxjs';
import { ComboboxOption } from '@grafana/ui';
import { DataSourceResponse, DEFAULT_QUERY, MyDataSourceOptions, MyQuery } from './types';

const RESOURCES_BASE = '/api/plugins/guilhermearpassos-sqlsights-app/resources';

const getOpts = async (query: string): Promise<Array<ComboboxOption<string>>> => {
  const response: Observable<FetchResponse<Array<{ label: string; value: string }>>> = getBackendSrv().fetch({
    url: `${RESOURCES_BASE}/datasource-options?${query}`,
  });
  const result = await lastValueFrom(response);
  return (result.data || []).map((o) => ({ label: o.label, value: o.value }));
};

function interpolateStringList(values: string[] | undefined, scopedVars: ScopedVars): string[] | undefined {
  if (!values || values.length === 0) {
    return values;
  }
  const tpl = getTemplateSrv();
  const out: string[] = [];
  for (const v of values) {
    const replaced = tpl.replace(v, scopedVars, 'csv');
    for (const part of replaced.split(',')) {
      const trimmed = part.trim();
      if (trimmed) {
        out.push(trimmed);
      }
    }
  }
  return out;
}

export function parseVariableQuery(raw: string): { kind: string; arg?: string } {
  const trimmed = raw.trim();
  const match = /^([a-zA-Z0-9_-]+)\s*(?:\(([^)]*)\))?$/.exec(trimmed);
  if (!match) {
    return { kind: trimmed.toLowerCase() };
  }
  return {
    kind: match[1].toLowerCase(),
    arg: match[2]?.trim() || undefined,
  };
}

export class DataSource extends DataSourceApi<MyQuery, MyDataSourceOptions> {
  baseUrl: string;

  constructor(instanceSettings: DataSourceInstanceSettings<MyDataSourceOptions>) {
    super(instanceSettings);
    this.baseUrl = instanceSettings.url!;
  }

  getDefaultQuery(_: CoreApp): Partial<MyQuery> {
    return DEFAULT_QUERY;
  }

  filterQuery(query: MyQuery): boolean {
    if (query.hide) {
      return false;
    }
    if (query.queryType === 'servers') {
      return true;
    }
    if (query.queryType === 'snapshot') {
      return Boolean(query.snapshotID);
    }
    return Boolean(query.database);
  }

  async query(options: DataQueryRequest<MyQuery>): Promise<DataQueryResponse> {
    const tpl = getTemplateSrv();
    const scopedVars = options.scopedVars || {};

    const preparedQueries = options.targets
      .filter((t) => this.filterQuery(t))
      .map((target) => {
        const interpolated: MyQuery = {
          ...target,
          queryType: target.queryType || 'chart',
          database: target.database ? tpl.replace(target.database, scopedVars) : target.database,
          databases: interpolateStringList(target.databases, scopedVars),
          issueTags: interpolateStringList(target.issueTags, scopedVars),
          snapshotID: target.snapshotID ? tpl.replace(target.snapshotID, scopedVars) : target.snapshotID,
          queryHash: target.queryHash ? tpl.replace(target.queryHash, scopedVars) : target.queryHash,
          waitEvent: target.waitEvent ? tpl.replace(target.waitEvent, scopedVars) : target.waitEvent,
          search: target.search ? tpl.replace(target.search, scopedVars) : target.search,
          datasource: {
            type: 'guilhermearpassos-sqlsights-datasource',
            uid: this.uid,
          },
        };
        return interpolated;
      });

    if (preparedQueries.length === 0) {
      return { data: [] };
    }

    const response = getBackendSrv().fetch<DataQueryResponse>({
      url: `${RESOURCES_BASE}/query`,
      method: 'POST',
      data: {
        queries: preparedQueries,
        range: options.range,
        intervalMs: options.intervalMs,
        maxDataPoints: options.maxDataPoints,
      },
    });

    const result = await lastValueFrom(response);
    return result.data;
  }

  /**
   * Enables Grafana Dashboard Query Variables:
   * - `servers` -> lists monitored SQL Server instances
   * - `databases($server)` -> lists databases observed on $server
   * - `snapshots($server)` -> lists recent snapshot IDs on $server
   * - `issue_tags` -> lists execution plan issue tags
   */
  async metricFindQuery(query: string, options?: any): Promise<MetricFindValue[]> {
    const tpl = getTemplateSrv();
    const replaced = tpl.replace(query || '', options?.scopedVars);
    const { kind, arg } = parseVariableQuery(replaced);

    const timeParams: Record<string, string> = {};
    if (options?.range?.from && options?.range?.to) {
      timeParams.start = options.range.from.toISOString();
      timeParams.end = options.range.to.toISOString();
    }

    let opts: Array<ComboboxOption<string>> = [];
    switch (kind) {
      case 'servers':
        opts = await this.fetchDropdownOptions('servers', timeParams);
        break;
      case 'databases':
        if (arg) {
          opts = await this.fetchDropdownOptions('server-databases', { ...timeParams, server: arg });
        } else {
          opts = await this.fetchDropdownOptions('servers', timeParams);
        }
        break;
      case 'snapshots':
        if (arg) {
          opts = await this.fetchDropdownOptions('snapshots', { ...timeParams, server: arg });
        }
        break;
      case 'issue_tags':
      case 'issue-tags':
      case 'issues':
        opts = await this.fetchDropdownOptions('issue-tags');
        break;
      default:
        opts = await this.fetchDropdownOptions('servers', timeParams);
    }

    return opts.map((o) => ({ text: o.label ?? o.value, value: o.value }));
  }

  private async fetchDropdownOptions(type: string, params?: Record<string, string>): Promise<Array<ComboboxOption<string>>> {
    const query = new URLSearchParams({ type, ...params });
    return getOpts(query.toString());
  }

  async getDropdownOptions(type: string, params?: Record<string, string>): Promise<Array<ComboboxOption<string>>> {
    try {
      return await this.fetchDropdownOptions(type, params);
    } catch (error) {
      console.error('Failed to fetch dropdown options:', error);
      return [];
    }
  }

  async getDatabaseOptions(): Promise<Array<ComboboxOption<string>>> {
    return this.getDropdownOptions('servers');
  }

  async getServerDatabaseOptions(server: string): Promise<Array<ComboboxOption<string>>> {
    if (!server) {
      return [];
    }
    return this.getDropdownOptions('server-databases', { server });
  }

  async getSnapshotOptions(server: string): Promise<Array<ComboboxOption<string>>> {
    if (!server) {
      return [];
    }
    return this.getDropdownOptions('snapshots', { server });
  }

  async request(url: string, params?: string) {
    const response = getBackendSrv().fetch<DataSourceResponse>({
      url: `${this.baseUrl}${url}${params?.length ? `?${params}` : ''}`,
    });
    return lastValueFrom(response);
  }

  async testDatasource() {
    const defaultErrorMessage = 'Cannot connect to Sqlsights app backend';
    try {
      const response = await lastValueFrom(
        getBackendSrv().fetch({
          url: `${RESOURCES_BASE}/datasource-options?type=servers`,
        })
      );
      if (response.status === 200) {
        return {
          status: 'success',
          message: 'Connected to Sqlsights API',
        };
      }
      return {
        status: 'error',
        message: response.statusText || defaultErrorMessage,
      };
    } catch (err) {
      let message = defaultErrorMessage;
      if (typeof err === 'string') {
        message = err;
      } else if (isFetchError(err)) {
        message = 'Fetch error: ' + (err.statusText || defaultErrorMessage);
        if (err.data?.error?.message) {
          message += ': ' + err.data.error.message;
        }
      }
      return {
        status: 'error',
        message,
      };
    }
  }
}
