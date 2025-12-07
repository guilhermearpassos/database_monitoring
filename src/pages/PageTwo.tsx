import React from 'react';
import { PluginPage } from '@grafana/runtime';

const PageTwo = () => {
    return (
        <PluginPage>
            <div>
                <h2>Data Sources Management</h2>
                <p>Configure and manage your SQL database connections.</p>

                <div style={{ marginTop: '24px' }}>
                    <h3>Available Data Sources</h3>
                    <p>Here you can add, configure, and test your SQL database connections that will be available as nested data sources within this plugin.</p>

                    {/* Add your datasource management UI here */}
                    <div style={{ padding: '16px', backgroundColor: '#f5f5f5', borderRadius: '4px', marginTop: '16px' }}>
                        <p><strong>Note:</strong> Data sources configured here will appear in Grafana's data source list as "SQL Insights DataSource".</p>
                    </div>
                </div>
            </div>
        </PluginPage>
    );
};

export default PageTwo;