import React from 'react';
import { Navigate, Route, Routes } from 'react-router-dom';
import { AppRootProps } from '@grafana/data';
import PagePlanAnalysis from "../../pages/PageTwo";
// import { ROUTES } from '../../constants';
const PageOne = React.lazy(() => import('../../pages/PageOne'));
// const PageTwo = React.lazy(() => import('../../pages/PageTwo'));
// const PageThree = React.lazy(() => import('../../pages/PageThree'));
// const PageFour = React.lazy(() => import('../../pages/PageFour'));

function App(props: AppRootProps) {
  return (
    <Routes>
      {/*<Route path={ROUTES.Two} element={<PageTwo />} />*/}
      {/*<Route path={`${ROUTES.Three}/:id?`} element={<PageThree />} />*/}

      {/*/!* Full-width page (this page will have no side navigation) *!/*/}
      {/*<Route path={ROUTES.Four} element={<PageFour />} />*/}

      {/* Default and fallback redirects */}
      <Route path="/" element={<Navigate to="/snapshots" replace />} />
      <Route path="*" element={<Navigate to="/snapshots" replace />} />

      {/* App pages */}
      <Route path="/snapshots" element={<PageOne />} />
      <Route path="/plan_analysis" element={<PagePlanAnalysis />} />
    </Routes>
  );
}

export default App;
