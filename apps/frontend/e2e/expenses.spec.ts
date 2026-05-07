import { test, expect, type Page } from '@playwright/test';
import { uniqueUser, register } from './helpers';

// Wait for the expenses page to finish loading, then open the create form and
// fill in a single-participant EQUAL-split expense.
async function createExpense(page: Page, desc: string, dollars: string, displayName: string) {
	// A freshly-registered user has no expenses, so "No expenses found." is
	// the reliable signal that the initial load completed.
	await expect(page.getByText('No expenses found.')).toBeVisible({ timeout: 15_000 });

	await page.getByRole('button', { name: '+ ADD EXPENSE' }).click();

	// Use placeholder selectors — more reliable than id which may not spread through
	// the Input component's {...rest} in this Svelte 5 version
	await page.getByPlaceholder('Dinner, Hotel, etc.').fill(desc);
	await page.getByPlaceholder('0.00').fill(dollars);

	// Participant select — native <select> whose first option is "Select user…"
	await page.locator('select').filter({ hasText: 'Select user…' }).selectOption({ label: displayName });
	await page.getByRole('button', { name: '+ ADD' }).click();
	await page.getByRole('button', { name: 'SAVE EXPENSE' }).click();
}

test('create EQUAL expense appears in list', async ({ page }) => {
	const user = uniqueUser();
	await register(page, user);
	await page.goto('/expenses');

	await createExpense(page, 'Team Lunch', '30.00', user.displayName);

	await expect(page.getByText('Team Lunch')).toBeVisible();
});

test('delete expense removes it from list', async ({ page }) => {
	const user = uniqueUser();
	await register(page, user);
	await page.goto('/expenses');

	await createExpense(page, 'Groceries', '50.00', user.displayName);
	await expect(page.getByText('Groceries')).toBeVisible();

	page.once('dialog', (dialog) => dialog.accept());
	await page.getByRole('button', { name: 'DEL' }).first().click();

	await expect(page.getByText('Groceries')).not.toBeVisible();
});

test('edit expense updates description', async ({ page }) => {
	const user = uniqueUser();
	await register(page, user);
	await page.goto('/expenses');

	await createExpense(page, 'Original Name', '20.00', user.displayName);
	await expect(page.getByText('Original Name')).toBeVisible();

	// Open inline edit form
	await page.getByRole('button', { name: 'EDIT' }).first().click();

	// The create form is not in the DOM when showForm=false, so the Description
	// placeholder uniquely targets the edit form's input
	await page.getByPlaceholder('Dinner, Hotel, etc.').fill('Updated Name');
	await page.getByRole('button', { name: 'SAVE' }).first().click();

	await expect(page.getByText('Updated Name')).toBeVisible();
	await expect(page.getByText('Original Name')).not.toBeVisible();
});
