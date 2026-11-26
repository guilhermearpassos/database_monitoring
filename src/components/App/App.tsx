import React, { Suspense } from 'react';
import { Navigate, Route, Routes } from 'react-router-dom';
import { AppRootProps } from '@grafana/data';
import { LoadingPlaceholder } from '@grafana/ui';
import PagePlanAnalysis from '../../pages/PageTwo';
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
        </Routes>
      </Suspense>
    </AppSettingsProvider>
  );
}

export default App;
