import React, { createContext, useContext, useMemo } from 'react';
import { TimeRange } from '@grafana/data';
import { getDataSourceSrv } from '@grafana/runtime';

export interface AppPluginJsonData {
  apiUrl?: string;
  tempoDatasourceUid?: string;
}

interface AppSettingsContextValue {
  apiUrl?: string;
  tempoDatasourceUid?: string;
  buildTempoExploreUrl: (traceId: string, timeRange?: TimeRange) => string | null;
}

const AppSettingsContext = createContext<AppSettingsContextValue>({
  buildTempoExploreUrl: () => null,
});

export function buildTempoExploreHref(tempoUid: string, traceId: string, timeRange?: TimeRange): string {
  const from = timeRange ? (typeof timeRange.raw.from === 'string' ? timeRange.raw.from : timeRange.from.toISOString()) : 'now-1h';
  const to = timeRange ? (typeof timeRange.raw.to === 'string' ? timeRange.raw.to : timeRange.to.toISOString()) : 'now';

  const paneState = {
    datasource: tempoUid,
    queries: [
      {
        refId: 'A',
        queryType: 'traceql',
        query: traceId,
        datasource: {
          type: 'tempo',
          uid: tempoUid,
        },
      },
    ],
    range: { from, to },
  };

  const params = new URLSearchParams({
    schemaVersion: '1',
    panes: JSON.stringify({ a: paneState }),
  });

  return `/explore?${params.toString()}`;
}

export const AppSettingsProvider: React.FC<{
  jsonData?: AppPluginJsonData;
  children: React.ReactNode;
}> = ({ jsonData, children }) => {
  const detectedTempoUid = useMemo<string | undefined>(() => {
    if (jsonData?.tempoDatasourceUid) {
      return undefined;
    }
    try {
      const dsSrv = getDataSourceSrv();
      if (!dsSrv || typeof dsSrv.getList !== 'function') {
        return undefined;
      }
      const tempoList = dsSrv.getList({ type: 'tempo' });
      if (tempoList && tempoList.length > 0) {
        return tempoList[0].uid;
      }
    } catch {
      // Ignore in test environments where DataSourceSrv is not initialized.
    }
    return undefined;
  }, [jsonData?.tempoDatasourceUid]);

  const effectiveTempoUid = jsonData?.tempoDatasourceUid || detectedTempoUid;

  const value = useMemo<AppSettingsContextValue>(
    () => ({
      apiUrl: jsonData?.apiUrl,
      tempoDatasourceUid: effectiveTempoUid,
      buildTempoExploreUrl: (traceId: string, timeRange?: TimeRange) => {
        if (!traceId || !effectiveTempoUid) {
          return null;
        }
        return buildTempoExploreHref(effectiveTempoUid, traceId, timeRange);
      },
    }),
    [jsonData?.apiUrl, effectiveTempoUid]
  );

  return <AppSettingsContext.Provider value={value}>{children}</AppSettingsContext.Provider>;
};

export const useAppSettings = () => useContext(AppSettingsContext);
