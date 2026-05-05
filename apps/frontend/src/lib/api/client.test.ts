import { describe, expect, test, beforeEach } from 'bun:test';
import { APIError, apiFetch } from './client';

function mockFetch(body: unknown, status: number): void {
	globalThis.fetch = (_: RequestInfo | URL, _init?: RequestInit) =>
		Promise.resolve(
			new Response(typeof body === 'string' ? body : JSON.stringify(body), {
				status,
				headers: { 'Content-Type': 'application/json' }
			})
		);
}

describe('APIError', () => {
	test('is an instance of Error with correct name and status', () => {
		const err = new APIError(404, 'not found');
		expect(err).toBeInstanceOf(Error);
		expect(err.name).toBe('APIError');
		expect(err.status).toBe(404);
		expect(err.message).toBe('not found');
	});
});

describe('apiFetch', () => {
	beforeEach(() => {
		mockFetch({ ok: true }, 200);
	});

	test('returns parsed JSON on 2xx', async () => {
		const result = await apiFetch<{ ok: boolean }>('/test');
		expect(result.ok).toBe(true);
	});

	test('throws APIError with correct status on 4xx', async () => {
		mockFetch({ error: 'not found' }, 404);
		await expect(apiFetch('/test')).rejects.toThrow(APIError);

		try {
			await apiFetch('/test');
		} catch (e) {
			expect(e).toBeInstanceOf(APIError);
			expect((e as APIError).status).toBe(404);
			expect((e as APIError).message).toBe('not found');
		}
	});

	test('throws APIError with correct status on 5xx', async () => {
		mockFetch({ error: 'internal server error' }, 500);

		try {
			await apiFetch('/test');
		} catch (e) {
			expect(e).toBeInstanceOf(APIError);
			expect((e as APIError).status).toBe(500);
		}
	});

	test('falls back to statusText when error body is not JSON', async () => {
		globalThis.fetch = () =>
			Promise.resolve(new Response('bad gateway', { status: 502, statusText: 'Bad Gateway' }));

		try {
			await apiFetch('/test');
		} catch (e) {
			expect(e).toBeInstanceOf(APIError);
			expect((e as APIError).status).toBe(502);
		}
	});

	test('clears auth store and throws on 401', async () => {
		mockFetch('', 401);

		// Pre-seed the auth store with a token
		const { authStore } = await import('../stores/auth');
		const { get } = await import('svelte/store');
		authStore.login('eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiJhbGljZSJ9.sig');

		await expect(apiFetch('/test')).rejects.toThrow(APIError);

		expect(get(authStore).token).toBeNull();
		expect(get(authStore).userID).toBeNull();
	});

	test('includes Authorization header when token is present', async () => {
		let capturedHeaders: Record<string, string> = {};
		globalThis.fetch = (_: RequestInfo | URL, init?: RequestInit) => {
			capturedHeaders = (init?.headers as Record<string, string>) ?? {};
			return Promise.resolve(new Response(JSON.stringify({}), { status: 200 }));
		};

		const { authStore } = await import('../stores/auth');
		authStore.login('eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiJhbGljZSJ9.sig');

		await apiFetch('/test');
		expect(capturedHeaders['Authorization']).toStartWith('Bearer ');
	});
});
