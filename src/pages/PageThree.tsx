import React from 'react';
import { PluginPage } from '@grafana/runtime';

const PageThree = () => {
    return (
        <PluginPage>
            <div>
                <h2>Analytics Dashboard</h2>
                <p>View database performance metrics and query analytics.</p>

                <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(250px, 1fr))', gap: '16px', marginTop: '24px' }}>
                    <div style={{ padding: '16px', border: '1px solid #ddd', borderRadius: '4px' }}>
                        <h4>Query Performance</h4>
                        <p>Average execution time, slow queries, and performance trends.</p>
                    </div>

                    <div style={{ padding: '16px', border: '1px solid #ddd', borderRadius: '4px' }}>
                        <h4>Database Health</h4>
                        <p>Connection status, error rates, and availability metrics.</p>
                    </div>

                    <div style={{ padding: '16px', border: '1px solid #ddd', borderRadius: '4px' }}>
                        <h4>Usage Statistics</h4>
                        <p>Query frequency, data source usage, and user activity.</p>
                    </div>
                </div>
            </div>
        </PluginPage>
    );
};

export default PageThree;