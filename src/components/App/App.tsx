import React, { Suspense } from 'react';
import { Navigate, Route, Routes } from 'react-router-dom';
import { AppRootProps } from '@grafana/data';
import { LoadingPlaceholder } from '@grafana/ui';
import PagePlanAnalysis from '../../pages/PageTwo';
import PageQueryMetrics from '../../pages/PageThree';
import PageDeadlocks from '../../pages/PageFour';
import { AppPluginJsonData, AppSettingsProvider } from '../../context/AppSettingsContext';

const PageOne = React.lazy(() => import('../../pages/PageOne'));

function App(props: AppRootProps<AppPluginJsonData>) {
  return (
    <AppSettingsProvider jsonData={props.meta?.jsonData}>
      <Suspense fallback={<LoadingPlaceholder text="Loading..." />}>
        <Routes>
          <Route path="/" element={<Navigate to="/snapshots" replace />} />
          <Route path="*" element={<Navigate to="/snapshots" replace />} />
          <Route path="/snapshots" element={<PageOne />} />
          <Route path="/plan_analysis" element={<PagePlanAnalysis />} />
          <Route path="/query_metrics" element={<PageQueryMetrics />} />
          <Route path="/deadlocks" element={<PageDeadlocks />} />
        </Routes>
      </Suspense>
    </AppSettingsProvider>
  );
}

export default App;
