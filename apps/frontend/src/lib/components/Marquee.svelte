<script lang="ts">
	let {
		text,
		speed = '18s',
		class: className = ''
	}: {
		text: string;
		speed?: string;
		class?: string;
	} = $props();
</script>

<div
	class="overflow-hidden bg-win-navy py-1 {className}"
	aria-hidden="true"
>
	<span class="marquee-inner text-win-yellow font-mono text-sm" style="--marquee-speed: {speed}">
		{text}
	</span>
</div>
