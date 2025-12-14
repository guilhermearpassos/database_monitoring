import React from 'react';
import { MemoryRouter } from 'react-router-dom';
import { AppRootProps, PluginType } from '@grafana/data';
import { render, waitFor } from '@testing-library/react';
import App from './App';

import 'jest-canvas-mock';
// At the top of App.test.tsx
jest.mock('@grafana/ui', () => ({
    ...jest.requireActual('@grafana/ui'),
    TimeRangePicker: () => <div data-testid="time-range-picker">Mocked TimeRangePicker</div>,
}));
describe('Components/App', () => {
  let props: AppRootProps;

  beforeEach(() => {
    jest.resetAllMocks();

    props = {
      basename: 'a/sample-app',
      meta: {
        id: 'sample-app',
        name: 'Sample App',
        type: PluginType.app,
        enabled: true,
        jsonData: {},
      },
      query: {},
      path: '',
      onNavChanged: jest.fn(),
    } as unknown as AppRootProps;
  });

  test('renders without an error"', async () => {
    const { queryByText } = render(
      <MemoryRouter>
        <App {...props} />
      </MemoryRouter>
    );

    // Application is lazy loaded, so we need to wait for the component and routes to be rendered
    await waitFor(() => expect(queryByText(/SQL Database Monitoring/i)).toBeInTheDocument(), { timeout: 2000 });
  });
});
