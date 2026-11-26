import React, { ChangeEvent, useMemo, useState } from 'react';
import { lastValueFrom } from 'rxjs';
import { css } from '@emotion/css';
import { AppPluginMeta, GrafanaTheme2, PluginConfigPageProps, PluginMeta } from '@grafana/data';
import { getBackendSrv, getDataSourceSrv } from '@grafana/runtime';
import { Button, Combobox, ComboboxOption, Field, FieldSet, Input, SecretInput, useStyles2 } from '@grafana/ui';
import { testIds } from '../testIds';

type AppPluginSettings = {
  apiUrl?: string;
  tempoDatasourceUid?: string;
};

type State = {
  apiUrl: string;
  isApiKeySet: boolean;
  apiKey: string;
  tempoDatasourceUid: string;
};

export interface AppConfigProps extends PluginConfigPageProps<AppPluginMeta<AppPluginSettings>> {}

const AppConfig = ({ plugin }: AppConfigProps) => {
  const s = useStyles2(getStyles);
  const { enabled, pinned, jsonData, secureJsonFields } = plugin.meta;
  const [state, setState] = useState<State>({
    apiUrl: jsonData?.apiUrl || '',
    apiKey: '',
    isApiKeySet: Boolean(secureJsonFields?.apiKey),
    tempoDatasourceUid: jsonData?.tempoDatasourceUid || '',
  });
  const tempoOptions = useMemo<Array<ComboboxOption<string>>>(() => {
    try {
      const dsSrv = getDataSourceSrv();
      if (!dsSrv || typeof dsSrv.getList !== 'function') {
        return [];
      }
      const list = dsSrv.getList({ type: 'tempo' }) || [];
      return list.map((ds) => ({
        label: `${ds.name} (${ds.uid})`,
        value: ds.uid,
        description: ds.url,
      }));
    } catch {
      // Ignore in test environments where DataSourceSrv is not initialized.
      return [];
    }
  }, []);

  const isSubmitDisabled = Boolean(!state.apiUrl);

  const onResetApiKey = () =>
    setState({
      ...state,
      apiKey: '',
      isApiKeySet: false,
    });

  const onChange = (event: ChangeEvent<HTMLInputElement>) => {
    setState({
      ...state,
      [event.target.name]: event.target.value.trim(),
    });
  };

  const onSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    if (isSubmitDisabled) {
      return;
    }

    updatePluginAndReload(plugin.meta.id, {
      enabled,
      pinned,
      jsonData: {
        apiUrl: state.apiUrl,
        tempoDatasourceUid: state.tempoDatasourceUid || undefined,
      },
      secureJsonData:
        state.isApiKeySet || !state.apiKey
          ? undefined
          : {
              apiKey: state.apiKey,
            },
    });
  };

  return (
    <form onSubmit={onSubmit}>
      <FieldSet label="API Settings">
        <Field label="API Key" description="Optional secret key for authenticating to your Sqlsights API">
          <SecretInput
            width={60}
            id="config-api-key"
            data-testid={testIds.appConfig.apiKey}
            name="apiKey"
            value={state.apiKey}
            isConfigured={state.isApiKeySet}
            placeholder={'Your secret API key'}
            onChange={onChange}
            onReset={onResetApiKey}
          />
        </Field>

        <Field label="API Url" description="gRPC host:port of your Sqlsights Querier service" className={s.marginTop}>
          <Input
            width={60}
            name="apiUrl"
            id="config-api-url"
            data-testid={testIds.appConfig.apiUrl}
            value={state.apiUrl}
            placeholder={`E.g.: sqlsights:9014`}
            onChange={onChange}
          />
        </Field>
      </FieldSet>

      <FieldSet label="Tracing Integration (Tempo)" className={s.marginTop}>
        <Field
          label="Tempo Datasource"
          description="Select a Grafana Tempo datasource to enable one-click trace navigation from SQL samples carrying a Trace ID (auto-detects the first Tempo datasource if left empty)."
        >
          <Combobox<string>
            width={60}
            options={tempoOptions}
            value={state.tempoDatasourceUid || null}
            onChange={(opt) => setState((prev) => ({ ...prev, tempoDatasourceUid: opt?.value || '' }))}
            placeholder={tempoOptions.length > 0 ? 'Auto-detect or choose Tempo datasource...' : 'No Tempo datasources found'}
            isClearable
          />
        </Field>
      </FieldSet>

      <div className={s.marginTop}>
        <Button type="submit" data-testid={testIds.appConfig.submit} disabled={isSubmitDisabled}>
          Save API settings
        </Button>
      </div>
    </form>
  );
};

export default AppConfig;

const getStyles = (theme: GrafanaTheme2) => ({
  colorWeak: css`
    color: ${theme.colors.text.secondary};
  `,
  marginTop: css`
    margin-top: ${theme.spacing(3)};
  `,
});

const updatePluginAndReload = async (pluginId: string, data: Partial<PluginMeta<AppPluginSettings>>) => {
  try {
    await updatePlugin(pluginId, data);
    window.location.reload();
  } catch (e) {
    console.error('Error while updating the plugin', e);
  }
};

const updatePlugin = async (pluginId: string, data: Partial<PluginMeta>) => {
  const response = await getBackendSrv().fetch({
    url: `/api/plugins/${pluginId}/settings`,
    method: 'POST',
    data,
  });

  return lastValueFrom(response);
};
