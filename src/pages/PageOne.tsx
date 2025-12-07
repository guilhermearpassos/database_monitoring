
import React, { useEffect, useState } from 'react';
import { PluginPage } from '@grafana/runtime';
import { getBackendSrv } from '@grafana/runtime';
import { lastValueFrom } from 'rxjs';
import {InteractiveTable} from "@grafana/ui";
const PageOne = () => {
    const [htmlContent, setHtmlContent] = useState<string>('Loading...');
    const [loading, setLoading] = useState<boolean>(true);
    const [error, setError] = useState<string | null>(null);

    // Function to fetch HTML content from backend
    const getMyCustomEndpoint = async () => {
        try {
            const response = await getBackendSrv().fetch({
                url: '/api/plugins/guilhermearpassos-sqlsights-app/resources/myCustomEndpoint',
            });
            // Get the response as text since it's HTML
            const textResponse = await lastValueFrom(response);
            return textResponse.data;
        } catch (err) {
            throw new Error(`Failed to fetch: ${err}`);
        }
    };

    // Fetch HTML content when component mounts
    useEffect(() => {
        const fetchContent = async () => {
            try {
                setLoading(true);
                const content = await getMyCustomEndpoint();
                setHtmlContent(content);
                setError(null);
            } catch (err) {
                setError(err instanceof Error ? err.message : 'Unknown error');
                setHtmlContent('');
            } finally {
                setLoading(false);
            }
        };

        fetchContent();
    }, []);
    return (
        <PluginPage>
            <div>
                <h2>SQL Insights Overview</h2>
                <p>Welcome to SQL Insights - your comprehensive database monitoring solution.</p>

                <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(300px, 1fr))', gap: '16px', marginTop: '24px' }}>
                    <div style={{ padding: '16px', border: '1px solid #ccc', borderRadius: '4px' }}>
                        <h3>Data Sources</h3>
                        <p>Manage your SQL database connections</p>
                    </div>

                    <div style={{ padding: '16px', border: '1px solid #ccc', borderRadius: '4px' }}>
                        <h3>Analytics</h3>
                        <p>View query performance and database insights</p>
                    </div>

                    <div style={{ padding: '16px', border: '1px solid #ccc', borderRadius: '4px' }}>
                        <h3>Reports</h3>
                        <p>Generate and view database reports</p>
                    </div>

                    <div style={{ padding: '16px', border: '1px solid #ccc', borderRadius: '4px' }}>
                        <h3>Custom Panels</h3>
                        <p>Use SQL Insights panels in your dashboards</p>
                    </div>

                    {/* Display the HTML content from backend */}
                    <div style={{ padding: '16px', border: '1px solid #ccc', borderRadius: '4px' }}>
                        <h3>Backend Content</h3>
                        {loading && <p>Loading backend content...</p>}
                        {error && <p style={{ color: 'red' }}>Error: {error}</p>}
                        {!loading && !error && (
                            <div
                                dangerouslySetInnerHTML={{ __html: htmlContent }}
                                style={{ border: '1px solid #eee', padding: '8px', backgroundColor: '#f9f9f9' }}
                            />
                        )}
                    </div>
                </div>
            </div>
        </PluginPage>
    );
};

export default PageOne;