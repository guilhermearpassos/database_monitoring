import React from 'react';
import { PluginPage } from '@grafana/runtime';

const PageFour = () => {
    return (
        <PluginPage>
            <div>
                <h2>Reports</h2>
                <p>Generate and manage database reports.</p>

                <div style={{ marginTop: '24px' }}>
                    <h3>Report Types</h3>

                    <div style={{ display: 'flex', flexDirection: 'column', gap: '16px', marginTop: '16px' }}>
                        <div style={{ padding: '16px', border: '1px solid #ddd', borderRadius: '4px' }}>
                            <h4>Performance Reports</h4>
                            <p>Detailed analysis of query performance, slow queries, and optimization suggestions.</p>
                        </div>

                        <div style={{ padding: '16px', border: '1px solid #ddd', borderRadius: '4px' }}>
                            <h4>Usage Reports</h4>
                            <p>Database usage statistics, connection patterns, and resource utilization.</p>
                        </div>

                        <div style={{ padding: '16px', border: '1px solid #ddd', borderRadius: '4px' }}>
                            <h4>Custom Reports</h4>
                            <p>Build custom reports with your own SQL queries and visualizations.</p>
                        </div>
                    </div>
                </div>
            </div>
        </PluginPage>
    );
};

export default PageFour;