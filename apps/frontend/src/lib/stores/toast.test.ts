import { describe, expect, test, beforeEach } from 'bun:test';
import { get } from 'svelte/store';
import { toastStore } from './toast';

beforeEach(() => {
	// Drain any lingering toasts between tests
	const current = get(toastStore);
	current.forEach((t) => toastStore.dismiss(t.id));
});

describe('toastStore.show', () => {
	test('adds a toast with correct type and message', () => {
		toastStore.show('success', 'It worked!');
		const toasts = get(toastStore);
		expect(toasts).toHaveLength(1);
		expect(toasts[0].type).toBe('success');
		expect(toasts[0].message).toBe('It worked!');
	});

	test('assigns unique ids to each toast', () => {
		toastStore.show('info', 'first');
		toastStore.show('info', 'second');
		const toasts = get(toastStore);
		expect(toasts[0].id).not.toBe(toasts[1].id);
	});
});

describe('toastStore.dismiss', () => {
	test('removes only the targeted toast', () => {
		toastStore.show('error', 'one');
		toastStore.show('error', 'two');
		const [a] = get(toastStore);
		toastStore.dismiss(a.id);
		const remaining = get(toastStore);
		expect(remaining).toHaveLength(1);
		expect(remaining[0].message).toBe('two');
	});
});

describe('toastStore convenience methods', () => {
	test('success() adds a success toast', () => {
		toastStore.success('saved');
		expect(get(toastStore)[0].type).toBe('success');
	});

	test('error() adds an error toast', () => {
		toastStore.error('failed');
		expect(get(toastStore)[0].type).toBe('error');
	});

	test('info() adds an info toast', () => {
		toastStore.info('note');
		expect(get(toastStore)[0].type).toBe('info');
	});
});
