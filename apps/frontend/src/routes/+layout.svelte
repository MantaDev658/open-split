<script lang="ts">
	import '../app.css';
	import { goto } from '$app/navigation';
	import { page } from '$app/stores';
	import { authStore } from '$lib/stores/auth';
	import type { Snippet } from 'svelte';

	let { children }: { children: Snippet } = $props();

	const PUBLIC = new Set(['/login', '/register']);

	$effect(() => {
		if (!$authStore.token && !PUBLIC.has($page.url.pathname)) {
			goto('/login');
		}
	});
</script>

{@render children()}
