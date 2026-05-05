<script lang="ts">
	import { getBalances, listExpenses } from '$lib/api/expenses';
	import HitCounter from '$lib/components/HitCounter.svelte';
	import Window from '$lib/components/Window.svelte';
	import { authStore } from '$lib/stores/auth';
	import { formatCents, formatDate } from '$lib/utils';
	import type { BalancesResponse, ExpenseItem } from '$lib/api/types';

	let balances = $state<BalancesResponse | null>(null);
	let expenses = $state<ExpenseItem[]>([]);
	let loading = $state(true);
	let error = $state('');

	const userID = $derived($authStore.userID ?? '');
	const userBalance = $derived(balances?.net_balances[userID] ?? 0);
	const owedToMe = $derived(Math.max(0, userBalance));
	const iOwe = $derived(Math.max(0, -userBalance));

	$effect(() => {
		async function load() {
			try {
				const [bal, exp] = await Promise.all([getBalances(), listExpenses(undefined, undefined, 5)]);
				balances = bal;
				expenses = exp.data;
			} catch {
				error = 'Failed to load dashboard data.';
			} finally {
				loading = false;
			}
		}
		load();
	});
</script>

<svelte:head>
	<title>Dashboard — Open Split</title>
</svelte:head>

{#if loading}
	<p class="font-system text-white text-sm animate-pulse">Loading…</p>
{:else if error}
	<p class="font-system text-win-red text-sm">{error}</p>
{:else}
	<!-- Balances + Settlements -->
	<div class="grid grid-cols-1 md:grid-cols-2 gap-4 mb-4">
		<!-- Hit counters -->
		<div class="flex flex-col gap-3">
			<HitCounter label="You Are Owed (cents)" value={owedToMe} />
			<HitCounter label="You Owe (cents)" value={iOwe} />
		</div>

		<!-- Suggested settlements -->
		<Window title="SUGGESTED SETTLEMENTS">
			{#if !balances?.suggested_settlements.length}
				<p class="font-system text-sm text-win-dark">All settled up! ✓</p>
			{:else}
				<table class="w-full font-system text-sm">
					<tbody>
						{#each balances.suggested_settlements as s, i}
							<tr class="{i % 2 === 0 ? 'bg-win-panel' : 'bg-white'} leading-6">
								<td class="px-2 font-bold">{s.From}</td>
								<td class="px-1 text-win-dark">→</td>
								<td class="px-2">{s.To}</td>
								<td class="px-2 text-right font-mono">{formatCents(s.Amount)}</td>
							</tr>
						{/each}
					</tbody>
				</table>
			{/if}
		</Window>
	</div>

	<!-- Construction stripe divider -->
	<div class="bg-construction h-6 mb-4" aria-hidden="true"></div>

	<!-- Recent expenses -->
	<Window title="RECENT EXPENSES">
		{#if !expenses.length}
			<p class="font-system text-sm text-win-dark">No expenses yet.</p>
		{:else}
			<table class="w-full font-system text-sm">
				<thead>
					<tr class="bg-win-navy text-white">
						<th class="px-2 py-1 text-left font-bold">Description</th>
						<th class="px-2 py-1 text-left font-bold">Payer</th>
						<th class="px-2 py-1 text-right font-bold">Amount</th>
						<th class="px-2 py-1 text-right font-bold">Date</th>
					</tr>
				</thead>
				<tbody>
					{#each expenses as exp, i}
						<tr class={i % 2 === 0 ? 'bg-win-panel' : 'bg-white'}>
							<td class="px-2 py-0.5">{exp.description}</td>
							<td class="px-2 py-0.5 text-win-dark">{exp.payer}</td>
							<td class="px-2 py-0.5 text-right font-mono">{formatCents(exp.total_cents)}</td>
							<td class="px-2 py-0.5 text-right text-win-dark">{formatDate(exp.created_at)}</td>
						</tr>
					{/each}
				</tbody>
			</table>
			<a href="/expenses" class="block text-xs text-win-accent underline mt-2 font-system">
				View all expenses →
			</a>
		{/if}
	</Window>
{/if}
