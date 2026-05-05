<script lang="ts">
	let {
		value = $bindable(''),
		options,
		placeholder = '',
		disabled = false,
		class: className = '',
		...rest
	}: {
		value?: string;
		options: { value: string; label: string }[];
		placeholder?: string;
		disabled?: boolean;
		class?: string;
		[key: string]: unknown;
	} = $props();
</script>

<select
	bind:value
	{disabled}
	{...rest}
	class="w-full px-2 py-1 bg-white text-black text-sm font-system
	       focus-visible:outline focus-visible:outline-2 focus-visible:outline-dotted focus-visible:outline-black
	       disabled:opacity-50 disabled:cursor-not-allowed
	       {className}"
	style="box-shadow: var(--bevel-in)"
>
	{#if placeholder}
		<option value="">{placeholder}</option>
	{/if}
	{#each options as opt}
		<option value={opt.value}>{opt.label}</option>
	{/each}
</select>
