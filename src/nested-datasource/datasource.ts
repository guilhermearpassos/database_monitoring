import {FetchResponse, getBackendSrv, getTemplateSrv, isFetchError} from '@grafana/runtime';
import {CoreApp, DataQueryRequest, DataQueryResponse, DataSourceApi, DataSourceInstanceSettings,} from '@grafana/data';

import {DataSourceResponse, DEFAULT_QUERY, MyDataSourceOptions, MyQuery} from './types';
import {lastValueFrom, Observable} from 'rxjs';
import {ComboboxOption} from "@grafana/ui";

// Function to fetch HTML content from backend
const getOpts: (query: string) => Promise<ComboboxOption[]> = async (query: string) => {
    try {
        let response: Observable<FetchResponse<ComboboxOption[]>>;
        response = await getBackendSrv().fetch({
            url: '/api/plugins/guilhermearpassos-sqlsights-app/resources/datasource-options?' + query,
        });
        // Get the response as text since it's HTML
        const textResponse = await lastValueFrom(response);
        return textResponse.data;
    } catch (err) {
        throw new Error(`Failed to fetch: ${err}`);
    }
};

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
        // if no query has been provided, prevent the query from being executed
        return !!query.database;
    }


    async query(options: DataQueryRequest<MyQuery>): Promise<DataQueryResponse> {
        // Use Grafana's built-in datasource proxy to call your backend's QueryData method
        const response = await getBackendSrv().fetch<DataQueryResponse>({
            url: '/api/plugins/guilhermearpassos-sqlsights-app/resources/query',
            method: 'POST',
            data: {
                queries: options.targets.map(target => {
                    console.log(target.snapshotID)
                    console.log(getTemplateSrv().getVariables())
                    target.snapshotID = getTemplateSrv().replace(target.snapshotID)
                    target.queryHash = getTemplateSrv().replace(target.queryHash)
                    console.log(target.snapshotID)
                    return {
                        ...target,
                        // Add datasource context
                        datasource: {
                            type: 'guilhermearpassos-sqlsights-datasource',
                            uid: this.uid,
                        },
                    }
                }),
                range: options.range,
                intervalMs: options.intervalMs,
                maxDataPoints: options.maxDataPoints,
            },
        });

        const result = await lastValueFrom(response);
        return result.data; // Extract the data from the FetchResponse
    }

    // Method to fetch dropdown options from backend
    async getDropdownOptions(type: string, params?: Record<string, string>): Promise<ComboboxOption[]> {
        const query = new URLSearchParams({type, ...params});

        try {
            const options = await getOpts(`${query.toString()}`);

            return options.map((option: ComboboxOption) => ({
                label: option.label,
                value: option.value,
            }));
        } catch (error) {
            console.error('Failed to fetch dropdown options:', error);
            return [];
        }
    }

    // Specific methods for different option types
    async getDatabaseOptions(): Promise<ComboboxOption[]> {
        return this.getDropdownOptions('databases');
    }

    async request(url: string, params?: string) {
        const response = getBackendSrv().fetch<DataSourceResponse>({
            url: `${this.baseUrl}${url}${params?.length ? `?${params}` : ''}`,
        });
        return lastValueFrom(response);
    }

    /**
     * Checks whether we can connect to the API.
     */
    async testDatasource() {
        const defaultErrorMessage = 'Cannot connect to API';

        try {
            const response = await this.request('/health');
            if (response.status === 200) {
                return {
                    status: 'success',
                    message: 'Success',
                };
            } else {
                return {
                    status: 'error',
                    message: response.statusText ? response.statusText : defaultErrorMessage,
                };
            }
        } catch (err) {
            let message = '';
            if (typeof err === 'string') {
                message = err;
            } else if (isFetchError(err)) {
                message = 'Fetch error: ' + (err.statusText ? err.statusText : defaultErrorMessage);
                if (err.data && err.data.error && err.data.error.code) {
                    message += ': ' + err.data.error.code + '. ' + err.data.error.message;
                }
            }
            return {
                status: 'error',
                message,
            };
        }
    }
}
