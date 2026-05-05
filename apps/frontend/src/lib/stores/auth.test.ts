import { describe, expect, test, beforeEach } from 'bun:test';
import { get } from 'svelte/store';
import { authStore } from './auth';

// JWT with payload {"sub":"alice"} (signature is fake — we never verify on the client)
const FAKE_TOKEN =
	'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiJhbGljZSJ9.fake-sig';

beforeEach(() => {
	authStore.logout();
});

describe('authStore.login', () => {
	test('sets token and decodes userID from JWT sub claim', () => {
		authStore.login(FAKE_TOKEN);
		const state = get(authStore);
		expect(state.token).toBe(FAKE_TOKEN);
		expect(state.userID).toBe('alice');
	});
});

describe('authStore.logout', () => {
	test('clears token and userID', () => {
		authStore.login(FAKE_TOKEN);
		authStore.logout();
		const state = get(authStore);
		expect(state.token).toBeNull();
		expect(state.userID).toBeNull();
	});
});

describe('authStore initial state', () => {
	test('starts with null token when localStorage is unavailable', () => {
		// In Bun test environment localStorage is undefined — store should start null
		const state = get(authStore);
		expect(state.token).toBeNull();
		expect(state.userID).toBeNull();
	});
});
