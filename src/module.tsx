import {AppPlugin, AppRootProps} from '@grafana/data';
import { AppConfigProps } from './components/AppConfig/AppConfig';

import React, {lazy, Suspense} from "react";
import {LoadingPlaceholder} from "@grafana/ui";

const LazyApp = lazy(() => import('./components/App/App'));
const LazyAppConfig = lazy(() => import('./components/AppConfig/AppConfig'));


const App = (props: AppRootProps) => (
    <Suspense fallback={<LoadingPlaceholder text="" />}>
        <LazyApp {...props} />
    </Suspense>
);

const AppConfig = (props: AppConfigProps) => (
    <Suspense fallback={<LoadingPlaceholder text="" />}>
        <LazyAppConfig {...props} />
    </Suspense>
);
// Main app plugin
export const plugin = new AppPlugin<{}>()
    .setRootPage(App)
    .addConfigPage({
        title: 'Configuration',
        icon: 'cog',
        body: AppConfig,
        id: 'configuration',
    });
