import pluginJson from './plugin.json';

export const PLUGIN_BASE_URL = `/a/${pluginJson.id}`;

export enum ROUTES {
  Snapshots = 'snapshots',
  PlanAnalysis = 'plan_analysis',
  QueryMetrics = 'query_metrics',
  Three = 'three',
  Four = 'four',
}
