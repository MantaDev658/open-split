import { authStore, getToken } from '../stores/auth';

const API_BASE = (import.meta.env?.VITE_API_BASE as string | undefined) ?? '/api';

export class APIError extends Error {
	constructor(
		public readonly status: number,
		message: string
	) {
		super(message);
		this.name = 'APIError';
	}
}

export async function apiFetch<T>(path: string, init: RequestInit = {}): Promise<T> {
	const token = getToken();

	const headers: Record<string, string> = {
		'Content-Type': 'application/json',
		...(init.headers as Record<string, string>)
	};

	if (token) {
		headers['Authorization'] = `Bearer ${token}`;
	}

	const res = await fetch(`${API_BASE}${path}`, { ...init, headers });

	if (res.status === 401) {
		authStore.logout();
		// Window guard: allows unit tests to run without a DOM
		if (typeof window !== 'undefined') {
			window.location.replace('/login');
		}
		throw new APIError(401, 'Session expired');
	}

	if (!res.ok) {
		let message = res.statusText;
		try {
			const body = await res.json();
			if (typeof body.error === 'string') message = body.error;
		} catch {
			// non-JSON error body — use statusText
		}
		throw new APIError(res.status, message);
	}

	return res.json() as Promise<T>;
}
