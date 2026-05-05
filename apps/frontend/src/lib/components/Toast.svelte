<script lang="ts">
	import { toastStore } from '$lib/stores/toast';

	const typeStyles = {
		success: 'border-l-4 border-win-green',
		error:   'border-l-4 border-win-red',
		info:    'border-l-4 border-win-navy'
	};

	const typeLabel = {
		success: '✓',
		error:   '⚠',
		info:    'i'
	};
</script>

<!-- Fixed system-tray position, bottom-right -->
<div class="fixed bottom-4 right-4 flex flex-col gap-2 z-50 pointer-events-none">
	{#each $toastStore as toast (toast.id)}
		<div
			class="pointer-events-auto flex items-start gap-2 bg-win95 px-3 py-2 min-w-48 max-w-72
			       font-system text-sm {typeStyles[toast.type]}"
			style="box-shadow: var(--bevel-out-deep)"
			role="alert"
		>
			<span class="font-bold shrink-0">{typeLabel[toast.type]}</span>
			<span class="flex-1">{toast.message}</span>
			<button
				class="shrink-0 text-win-dark hover:text-black leading-none"
				onclick={() => toastStore.dismiss(toast.id)}
				aria-label="Dismiss"
			>✕</button>
		</div>
	{/each}
</div>
