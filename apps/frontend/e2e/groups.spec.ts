import { test, expect } from '@playwright/test';
import { uniqueUser, register } from './helpers';

test('empty groups page shows no groups message', async ({ page }) => {
	const user = uniqueUser();
	await register(page, user);
	await page.goto('/groups');
	await expect(page.getByText('No groups yet. Create one above.')).toBeVisible({ timeout: 15_000 });
});

test('create group appears in list', async ({ page }) => {
	const user = uniqueUser();
	await register(page, user);
	await page.goto('/groups');
	await expect(page.getByText('No groups yet. Create one above.')).toBeVisible({ timeout: 15_000 });
	await page.getByRole('button', { name: '+ CREATE GROUP' }).click();
	await page.fill('[placeholder="Group name…"]', 'E2E Test Group');
	await page.getByRole('button', { name: 'CREATE' }).click();
	await expect(page.getByText('E2E Test Group')).toBeVisible();
});

test('rename group updates the name', async ({ page }) => {
	const user = uniqueUser();
	await register(page, user);
	await page.goto('/groups');
	await expect(page.getByText('No groups yet. Create one above.')).toBeVisible({ timeout: 15_000 });

	// Create
	await page.getByRole('button', { name: '+ CREATE GROUP' }).click();
	await page.fill('[placeholder="Group name…"]', 'Old Name');
	await page.getByRole('button', { name: 'CREATE' }).click();
	await expect(page.getByText('Old Name')).toBeVisible();

	// Open group detail
	await page.getByText('Old Name').first().click();
	await page.waitForURL(/\/groups\/.+/);

	// Rename
	await page.getByRole('button', { name: 'RENAME GROUP' }).click();
	await page.fill('[placeholder="New name…"]', 'Renamed Group');
	await page.getByRole('button', { name: 'SAVE' }).click();

	// After save, the Window title updates to the new name
	await expect(page.getByText('Renamed Group')).toBeVisible();
});

test('delete group redirects to /groups and removes it from list', async ({ page }) => {
	const user = uniqueUser();
	await register(page, user);
	await page.goto('/groups');
	await expect(page.getByText('No groups yet. Create one above.')).toBeVisible({ timeout: 15_000 });

	// Create
	await page.getByRole('button', { name: '+ CREATE GROUP' }).click();
	await page.fill('[placeholder="Group name…"]', 'To Be Deleted');
	await page.getByRole('button', { name: 'CREATE' }).click();
	await expect(page.getByText('To Be Deleted')).toBeVisible();

	// Open group detail
	await page.getByText('To Be Deleted').first().click();
	await page.waitForURL(/\/groups\/.+/);

	// Accept the confirm dialog and delete
	page.once('dialog', (dialog) => dialog.accept());
	await page.getByRole('button', { name: 'DELETE GROUP' }).click();

	await page.waitForURL('/groups');
	await expect(page).toHaveURL('/groups');
	await expect(page.getByText('To Be Deleted')).not.toBeVisible();
});
