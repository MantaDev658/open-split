<script lang="ts">
	import type { Snippet } from 'svelte';

	type Variant = 'default' | 'primary' | 'danger' | 'success';

	let {
		variant = 'default',
		type = 'button',
		disabled = false,
		onclick,
		children,
		class: className = ''
	}: {
		variant?: Variant;
		type?: 'button' | 'submit' | 'reset';
		disabled?: boolean;
		onclick?: (e: MouseEvent) => void;
		children: Snippet;
		class?: string;
	} = $props();

	let pressed = $state(false);

	const bg: Record<Variant, string> = {
		default: 'bg-win95 text-black',
		primary: 'bg-win-navy text-white',
		danger:  'bg-win-red  text-white',
		success: 'bg-win-green text-black'
	};
</script>

<button
	{type}
	{disabled}
	{onclick}
	onmousedown={() => (pressed = true)}
	onmouseup={() => (pressed = false)}
	onmouseleave={() => (pressed = false)}
	class="px-4 py-1 font-bold uppercase text-sm font-system
	       focus-visible:outline focus-visible:outline-2 focus-visible:outline-dotted focus-visible:outline-black
	       disabled:opacity-50 disabled:cursor-not-allowed
	       {bg[variant]} {className}"
	style="box-shadow: {pressed ? 'var(--bevel-in)' : 'var(--bevel-out)'}; transform: {pressed
		? 'translate(1px,1px)'
		: 'none'}"
>
	{@render children()}
</button>
