<script lang="ts">
	let {
		text = 'NEW!',
		variant = 'red',
		class: className = ''
	}: {
		text?: string;
		variant?: 'red' | 'green' | 'yellow' | 'navy';
		class?: string;
	} = $props();

	const colors: Record<string, string> = {
		red:    'bg-win-red   text-white',
		green:  'bg-win-green text-black',
		yellow: 'bg-win-yellow text-black',
		navy:   'bg-win-navy  text-white'
	};
</script>

<span
	class="pulse-glow inline-block text-xs font-bold uppercase px-1.5 py-0.5 font-system
	       {colors[variant]} {className}"
>
	{text}
</span>
