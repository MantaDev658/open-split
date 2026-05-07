import { test, expect } from '@playwright/test';
import { uniqueUser, register, registerViaApi } from './helpers';

test('settle up records a payment and shows success toast', async ({ page }) => {
	// Create a second user directly via the API so they appear in the recipients list
	const recipient = uniqueUser();
	await registerViaApi(recipient);

	// Register and log in as the payer via the UI
	const payer = uniqueUser();
	await register(page, payer);

	await page.goto('/settle');

	// Wait for the form to finish loading — the select is disabled while loading
	await expect(page.locator('#settle-receiver')).toBeEnabled({ timeout: 15_000 });

	// Select the recipient
	await page.locator('#settle-receiver').selectOption({ label: recipient.displayName });

	// Enter an amount
	await page.fill('#settle-amount', '25.00');

	// Submit
	await page.getByRole('button', { name: 'SETTLE UP' }).click();

	// Confirm success toast
	await expect(page.getByRole('alert').getByText('Settlement recorded!')).toBeVisible();
});

test('settle form validates missing recipient', async ({ page }) => {
	const user = uniqueUser();
	await register(page, user);
	await page.goto('/settle');

	// Wait for the form to finish loading — the select is disabled while loading
	await expect(page.locator('#settle-amount')).toBeEnabled({ timeout: 15_000 });

	// Submit without selecting a recipient
	await page.fill('#settle-amount', '10.00');
	await page.getByRole('button', { name: 'SETTLE UP' }).click();

	await expect(page.getByRole('alert').getByText('Select a recipient.')).toBeVisible();
});
