import {DataSourceJsonData} from '@grafana/data';
import {DataQuery} from '@grafana/schema';

export interface MyQuery extends DataQuery {
    database?: string;
    snapshotID?: string;
    // Add other query parameters as needed
}

export const DEFAULT_QUERY: Partial<MyQuery> = {
    database: 'default',
};

export interface DataPoint {
    Time: number;
    Value: number;
}

export interface DataSourceResponse {
    datapoints: DataPoint[];
}

/**
 * These are options configured for each DataSource instance
 */
export interface MyDataSourceOptions extends DataSourceJsonData {
    path?: string;
}

/**
 * Value that is used in the backend, but never sent over HTTP to the frontend
 */
export interface MySecureJsonData {
    apiKey?: string;
}
