import { writable } from 'svelte/store';

export type ToastType = 'success' | 'error' | 'info';

export interface Toast {
	id: number;
	type: ToastType;
	message: string;
}

let nextId = 0;

function createToastStore() {
	const { subscribe, update } = writable<Toast[]>([]);

	function dismiss(id: number) {
		update((toasts) => toasts.filter((t) => t.id !== id));
	}

	function show(type: ToastType, message: string, durationMs = 3500) {
		const id = ++nextId;
		update((toasts) => [...toasts, { id, type, message }]);
		setTimeout(() => dismiss(id), durationMs);
	}

	return {
		subscribe,
		show,
		dismiss,
		success: (msg: string) => show('success', msg),
		error: (msg: string) => show('error', msg),
		info: (msg: string) => show('info', msg)
	};
}

export const toastStore = createToastStore();
