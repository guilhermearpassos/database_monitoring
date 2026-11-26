import { test, expect } from './fixtures';
import { ROUTES } from '../src/constants';

test.describe('navigating app', () => {
  test('snapshots page should render successfully', async ({ gotoPage, page }) => {
    await gotoPage(`/${ROUTES.Snapshots}`);
    await expect(page.getByText('SQL Database Monitoring')).toBeVisible();
  });

  test('plan analysis page should render successfully', async ({ gotoPage, page }) => {
    await gotoPage(`/${ROUTES.PlanAnalysis}`);
    await expect(page.getByText('Plan Analysis Explorer')).toBeVisible();
  });

  test('query metrics page should render successfully', async ({ gotoPage, page }) => {
    await gotoPage(`/${ROUTES.QueryMetrics}`);
    await expect(page.getByText('Query Metrics & Workload Stats')).toBeVisible();
  });
});
