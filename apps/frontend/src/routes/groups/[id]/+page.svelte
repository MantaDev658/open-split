<script lang="ts">
	import { page } from '$app/stores';
	import { APIError } from '$lib/api/client';
	import { deleteExpense, listExpenses } from '$lib/api/expenses';
	import {
		addGroupMember,
		getGroupActivity,
		listGroups,
		removeGroupMember
	} from '$lib/api/groups';
	import type { AuditLog, ExpenseItem, Group, User } from '$lib/api/types';
	import { listUsers } from '$lib/api/users';
	import Button from '$lib/components/Button.svelte';
	import HRule from '$lib/components/HRule.svelte';
	import Select from '$lib/components/Select.svelte';
	import Window from '$lib/components/Window.svelte';
	import { toastStore } from '$lib/stores/toast';
	import { formatCents, formatDate } from '$lib/utils';

	type Tab = 'members' | 'expenses' | 'activity';

	// ── Route ────────────────────────────────────────────────────────
	const groupID = $derived($page.params.id ?? '');

	// ── Data ─────────────────────────────────────────────────────────
	let group = $state<Group | null>(null);
	let allUsers = $state<User[]>([]);
	let expenses = $state<ExpenseItem[]>([]);
	let activity = $state<AuditLog[]>([]);
	let activityCursor = $state('');
	let activityCursorStack = $state<string[]>([]);

	// ── UI ───────────────────────────────────────────────────────────
	let loading = $state(true);
	let tab = $state<Tab>('members');
	let addMemberID = $state('');
	let addingMember = $state(false);

	// ── Derived ──────────────────────────────────────────────────────
	const userByID = $derived(Object.fromEntries(allUsers.map((u) => [u.ID, u])));

	const nonMembers = $derived(
		allUsers.filter((u) => !group?.Members.includes(u.ID))
			.map((u) => ({ value: u.ID, label: u.DisplayName }))
	);

	// ── Load ─────────────────────────────────────────────────────────
	$effect(() => {
		const id = groupID;
		loading = true;
		Promise.all([listGroups(), listUsers()])
			.then(([groups, users]) => {
				group = groups.find((g) => g.ID === id) ?? null;
				allUsers = users;
			})
			.catch(() => toastStore.error('Failed to load group.'))
			.finally(() => (loading = false));
	});

	$effect(() => {
		if (tab === 'expenses' && groupID) loadExpenses();
		if (tab === 'activity' && groupID) loadActivity();
	});

	async function loadExpenses() {
		try {
			const result = await listExpenses(groupID, undefined, 20);
			expenses = result.data;
		} catch {
			toastStore.error('Failed to load expenses.');
		}
	}

	async function loadActivity(cursor = '') {
		try {
			const result = await getGroupActivity(groupID, cursor || undefined, 20);
			activity = result.data;
			activityCursor = result.next_cursor;
		} catch {
			toastStore.error('Failed to load activity.');
		}
	}

	// ── Members ───────────────────────────────────────────────────────
	async function handleAddMember() {
		if (!addMemberID || addingMember) return;
		addingMember = true;
		try {
			await addGroupMember(groupID, addMemberID);
			toastStore.success('Member added.');
			addMemberID = '';
			const groups = await listGroups();
			group = groups.find((g) => g.ID === groupID) ?? null;
		} catch (err) {
			toastStore.error(err instanceof APIError ? err.message : 'Failed to add member.');
		} finally {
			addingMember = false;
		}
	}

	async function handleRemoveMember(userID: string) {
		if (!confirm(`Remove ${userByID[userID]?.DisplayName ?? userID} from group?`)) return;
		try {
			await removeGroupMember(groupID, userID);
			toastStore.success('Member removed.');
			const groups = await listGroups();
			group = groups.find((g) => g.ID === groupID) ?? null;
		} catch (err) {
			toastStore.error(err instanceof APIError ? err.message : 'Failed to remove member.');
		}
	}

	// ── Expenses ──────────────────────────────────────────────────────
	async function handleDeleteExpense(id: string) {
		if (!confirm('Delete this expense?')) return;
		try {
			await deleteExpense(id);
			await loadExpenses();
		} catch (err) {
			toastStore.error(err instanceof APIError ? err.message : 'Failed to delete expense.');
		}
	}

	// ── Activity pagination ───────────────────────────────────────────
	function activityNext() {
		activityCursorStack = [...activityCursorStack, activityCursor];
		loadActivity(activityCursor);
	}

	function activityPrev() {
		const stack = [...activityCursorStack];
		stack.pop();
		const cursor = stack.at(-1) ?? '';
		activityCursorStack = stack;
		loadActivity(cursor);
	}
</script>

<svelte:head>
	<title>{group?.Name ?? 'Group'} — Open Split</title>
</svelte:head>

{#if loading}
	<p class="font-system text-white text-sm animate-pulse">Loading…</p>
{:else if !group}
	<p class="font-system text-white text-sm">Group not found or you are not a member.</p>
{:else}
	<Window title={group.Name}>
		<!-- Tab bar -->
		<div class="flex gap-0 -mx-4 -mt-4 mb-4 overflow-x-auto">
			{#each (['members', 'expenses', 'activity'] as Tab[]) as t}
				<button
					class="px-4 py-1.5 text-xs font-bold uppercase font-system shrink-0
					       {tab === t ? 'bg-win95 text-black' : 'bg-win-dark text-white hover:bg-win95 hover:text-black'}"
					style="box-shadow: {tab === t ? 'var(--bevel-out)' : 'var(--bevel-in)'}"
					onclick={() => (tab = t)}
				>
					{t}
				</button>
			{/each}
		</div>

		<!-- Members tab -->
		{#if tab === 'members'}
			<div class="flex flex-col gap-2 font-system text-sm">
				{#each group.Members as memberID, i}
					<div
						class="flex items-center justify-between px-2 py-1
						       {i % 2 === 0 ? 'bg-win-panel' : 'bg-white'}"
					>
						<span>{userByID[memberID]?.DisplayName ?? memberID}</span>
						<Button variant="danger" onclick={() => handleRemoveMember(memberID)}>REMOVE</Button>
					</div>
				{/each}

				<HRule />

				<div class="flex gap-2 mt-1">
					<Select
						bind:value={addMemberID}
						placeholder="Add member…"
						options={nonMembers}
						class="flex-1"
					/>
					<Button variant="success" onclick={handleAddMember} disabled={!addMemberID || addingMember}>
						{addingMember ? '…' : '+ ADD'}
					</Button>
				</div>
			</div>

		<!-- Expenses tab -->
		{:else if tab === 'expenses'}
			{#if !expenses.length}
				<p class="font-system text-sm text-win-dark">No expenses in this group yet.</p>
			{:else}
				<div class="overflow-x-auto">
					<table class="w-full font-system text-sm">
						<thead>
							<tr class="bg-win-navy text-white">
								<th class="px-2 py-1 text-left font-bold">Description</th>
								<th class="px-2 py-1 text-left font-bold">Payer</th>
								<th class="px-2 py-1 text-right font-bold">Amount</th>
								<th class="px-2 py-1 text-right font-bold">Date</th>
								<th class="px-2 py-1"></th>
							</tr>
						</thead>
						<tbody>
							{#each expenses as exp, i}
								<tr class={i % 2 === 0 ? 'bg-win-panel' : 'bg-white'}>
									<td class="px-2 py-0.5">{exp.description}</td>
									<td class="px-2 py-0.5 text-win-dark">{exp.payer}</td>
									<td class="px-2 py-0.5 text-right font-mono">{formatCents(exp.total_cents)}</td>
									<td class="px-2 py-0.5 text-right text-win-dark">{formatDate(exp.created_at)}</td>
									<td class="px-2 py-0.5 text-right">
										<Button variant="danger" onclick={() => handleDeleteExpense(exp.id)}>DEL</Button>
									</td>
								</tr>
							{/each}
						</tbody>
					</table>
				</div>
			{/if}

		<!-- Activity tab -->
		{:else if tab === 'activity'}
			{#if !activity.length}
				<p class="font-system text-sm text-win-dark">No activity yet.</p>
			{:else}
				<div class="flex flex-col gap-1 font-mono text-xs">
					{#each activity as log, i}
						<div class="px-2 py-1 {i % 2 === 0 ? 'bg-win-panel' : 'bg-white'}">
							<span class="text-win-dark">{formatDate(log.created_at)}</span>
							<span class="font-bold mx-2">{log.action}</span>
							<span class="text-win-dark">{log.user_id}</span>
							{#if log.details}
								<span class="ml-2 italic">{log.details}</span>
							{/if}
						</div>
					{/each}
				</div>
				<div class="flex gap-2 mt-3">
					<Button onclick={activityPrev} disabled={!activityCursorStack.length}>◀ PREV</Button>
					<Button onclick={activityNext} disabled={!activityCursor}>NEXT ▶</Button>
				</div>
			{/if}
		{/if}
	</Window>
{/if}
