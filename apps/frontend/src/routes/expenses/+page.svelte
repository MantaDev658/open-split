<script lang="ts">
	import { APIError } from '$lib/api/client';
	import { createExpense, deleteExpense, listExpenses } from '$lib/api/expenses';
	import { listGroups } from '$lib/api/groups';
	import type { ExpenseItem, Group, SplitInput, SplitType, User } from '$lib/api/types';
	import { listUsers } from '$lib/api/users';
	import Button from '$lib/components/Button.svelte';
	import HRule from '$lib/components/HRule.svelte';
	import Input from '$lib/components/Input.svelte';
	import Select from '$lib/components/Select.svelte';
	import Window from '$lib/components/Window.svelte';
	import { formatCents, formatDate } from '$lib/utils';

	// ── Data ────────────────────────────────────────────────────────
	let expenses = $state<ExpenseItem[]>([]);
	let groups = $state<Group[]>([]);
	let users = $state<User[]>([]);
	let nextCursor = $state('');
	let cursorStack = $state<string[]>([]);

	// ── UI state ────────────────────────────────────────────────────
	let loading = $state(true);
	let error = $state('');
	let groupFilter = $state('');
	let showForm = $state(false);

	// ── Create form ─────────────────────────────────────────────────
	let desc = $state('');
	let totalDollars = $state('');
	let splitType = $state<SplitType>('EVEN');
	let formGroupId = $state('');
	let participants = $state<{ userID: string; value: string }[]>([]);
	let addUserID = $state('');
	let formError = $state('');
	let submitting = $state(false);

	// ── Derived ─────────────────────────────────────────────────────
	const userByID = $derived(Object.fromEntries(users.map((u) => [u.ID, u])));

	const groupOptions = $derived([
		{ value: '', label: 'All Expenses' },
		...groups.map((g) => ({ value: g.ID, label: g.Name }))
	]);

	const userOptions = $derived(
		users
			.filter((u) => !participants.some((p) => p.userID === u.ID))
			.map((u) => ({ value: u.ID, label: u.DisplayName }))
	);

	const splitValueMeta = $derived(
		splitType === 'EXACT'
			? { label: 'Amount ($)', placeholder: '0.00' }
			: splitType === 'PERCENT'
				? { label: 'Percent (%)', placeholder: '0' }
				: splitType === 'SHARES'
					? { label: 'Shares', placeholder: '1' }
					: null
	);

	// ── Data loading ─────────────────────────────────────────────────
	$effect(() => {
		loadAll();
	});

	async function loadAll() {
		loading = true;
		error = '';
		try {
			const [exp, grps, usrs] = await Promise.all([
				listExpenses(undefined, undefined, 20),
				listGroups(),
				listUsers()
			]);
			expenses = exp.data;
			nextCursor = exp.next_cursor;
			groups = grps;
			users = usrs;
		} catch {
			error = 'Failed to load expenses.';
		} finally {
			loading = false;
		}
	}

	async function loadPage(cursor = '') {
		loading = true;
		error = '';
		try {
			const result = await listExpenses(groupFilter || undefined, cursor || undefined, 20);
			expenses = result.data;
			nextCursor = result.next_cursor;
		} catch {
			error = 'Failed to load expenses.';
		} finally {
			loading = false;
		}
	}

	function applyFilter() {
		cursorStack = [];
		loadPage();
	}

	function goNext() {
		cursorStack = [...cursorStack, nextCursor];
		loadPage(nextCursor);
	}

	function goPrev() {
		const stack = [...cursorStack];
		stack.pop();
		const cursor = stack.at(-1) ?? '';
		cursorStack = stack;
		loadPage(cursor);
	}

	// ── Create expense ───────────────────────────────────────────────
	function addParticipant() {
		if (!addUserID) return;
		participants = [...participants, { userID: addUserID, value: '' }];
		addUserID = '';
	}

	function removeParticipant(userID: string) {
		participants = participants.filter((p) => p.userID !== userID);
	}

	function updateParticipantValue(userID: string, value: string) {
		participants = participants.map((p) => (p.userID === userID ? { ...p, value } : p));
	}

	function resetForm() {
		desc = '';
		totalDollars = '';
		splitType = 'EVEN';
		formGroupId = '';
		participants = [];
		addUserID = '';
		formError = '';
	}

	async function handleCreate(e: SubmitEvent) {
		e.preventDefault();
		if (submitting) return;

		const totalCents = Math.round(parseFloat(totalDollars) * 100);
		if (!desc.trim()) { formError = 'Description is required.'; return; }
		if (isNaN(totalCents) || totalCents <= 0) { formError = 'Enter a valid amount.'; return; }
		if (participants.length === 0) { formError = 'Add at least one participant.'; return; }

		const splits: SplitInput[] = participants.map((p) => {
			if (splitType === 'EVEN') return { user_id: p.userID };
			const num = parseFloat(p.value);
			const value = splitType === 'EXACT' ? Math.round(num * 100) : Math.round(num);
			return { user_id: p.userID, value };
		});

		submitting = true;
		formError = '';
		try {
			await createExpense({
				description: desc.trim(),
				total_cents: totalCents,
				split_type: splitType,
				splits,
				...(formGroupId ? { group_id: formGroupId } : {})
			});
			resetForm();
			showForm = false;
			cursorStack = [];
			loadPage();
		} catch (err) {
			formError = err instanceof APIError ? err.message : 'Failed to create expense.';
		} finally {
			submitting = false;
		}
	}

	// ── Delete expense ────────────────────────────────────────────────
	async function handleDelete(id: string) {
		if (!confirm('Delete this expense?')) return;
		try {
			await deleteExpense(id);
			const cursor = cursorStack.at(-1) ?? '';
			loadPage(cursor);
		} catch (err) {
			error = err instanceof APIError ? err.message : 'Failed to delete expense.';
		}
	}
</script>

<svelte:head>
	<title>Expenses — Open Split</title>
</svelte:head>

<Window title="EXPENSE LEDGER">
	<!-- Toolbar -->
	<div class="flex items-center gap-2 mb-3 flex-wrap">
		<Button
			onclick={() => {
				showForm = !showForm;
				if (!showForm) resetForm();
			}}
			variant={showForm ? 'default' : 'primary'}
		>
			{showForm ? 'CANCEL' : '+ ADD EXPENSE'}
		</Button>

		<div class="flex items-center gap-1 ml-auto">
			<label class="text-xs font-system font-bold shrink-0" for="group-filter">GROUP:</label>
			<Select
				id="group-filter"
				bind:value={groupFilter}
				options={groupOptions}
				class="w-40"
				onchange={applyFilter}
			/>
		</div>
	</div>

	<!-- Create form -->
	{#if showForm}
		<div class="mb-4">
			<HRule />
			<form class="mt-3 flex flex-col gap-3 font-system" onsubmit={handleCreate}>
				<div class="grid grid-cols-1 sm:grid-cols-2 gap-3">
					<div class="flex flex-col gap-1">
						<label class="text-xs font-bold" for="exp-desc">Description</label>
						<Input id="exp-desc" bind:value={desc} placeholder="Dinner, Hotel, etc." />
					</div>
					<div class="flex flex-col gap-1">
						<label class="text-xs font-bold" for="exp-total">Total ($)</label>
						<Input id="exp-total" type="number" min="0.01" step="0.01" bind:value={totalDollars} placeholder="0.00" />
					</div>
					<div class="flex flex-col gap-1">
						<label class="text-xs font-bold" for="exp-split">Split Type</label>
						<Select
							id="exp-split"
							bind:value={splitType}
							options={[
								{ value: 'EVEN', label: 'Even' },
								{ value: 'EXACT', label: 'Exact amounts' },
								{ value: 'PERCENT', label: 'Percentages' },
								{ value: 'SHARES', label: 'Shares' }
							]}
						/>
					</div>
					<div class="flex flex-col gap-1">
						<label class="text-xs font-bold" for="exp-group">Group (optional)</label>
						<Select
							id="exp-group"
							bind:value={formGroupId}
							placeholder="No group"
							options={groups.map((g) => ({ value: g.ID, label: g.Name }))}
						/>
					</div>
				</div>

				<!-- Participants -->
				<div class="flex flex-col gap-1">
					<span class="text-xs font-bold">Participants</span>
					{#each participants as p}
						<div class="flex items-center gap-2">
							<span class="text-xs flex-1 truncate">
								{userByID[p.userID]?.DisplayName ?? p.userID}
							</span>
							{#if splitValueMeta}
								<Input
									type="number"
									min="0"
									step={splitType === 'EXACT' ? '0.01' : '1'}
									placeholder={splitValueMeta.placeholder}
									value={p.value}
									oninput={(e: Event) => updateParticipantValue(p.userID, (e.currentTarget as HTMLInputElement).value)}
									class="w-24"
								/>
								<span class="text-xs text-win-dark shrink-0">{splitValueMeta.label}</span>
							{/if}
							<Button variant="danger" onclick={() => removeParticipant(p.userID)}>✕</Button>
						</div>
					{/each}
					<div class="flex gap-2 mt-1">
						<Select
							bind:value={addUserID}
							placeholder="Select user…"
							options={userOptions}
							class="flex-1"
						/>
						<Button onclick={addParticipant} disabled={!addUserID}>+ ADD</Button>
					</div>
				</div>

				{#if formError}
					<p class="text-xs text-win-red px-2 py-1" style="box-shadow: var(--bevel-in)">
						⚠ {formError}
					</p>
				{/if}

				<div class="flex gap-2">
					<Button type="submit" variant="success" disabled={submitting}>
						{submitting ? 'SAVING…' : 'SAVE EXPENSE'}
					</Button>
				</div>
			</form>
			<HRule class="mt-3" />
		</div>
	{/if}

	<!-- Error -->
	{#if error}
		<p class="text-xs text-win-red mb-2 font-system">⚠ {error}</p>
	{/if}

	<!-- Table -->
	{#if loading}
		<p class="font-system text-sm text-win-dark animate-pulse">Loading…</p>
	{:else if !expenses.length}
		<p class="font-system text-sm text-win-dark">No expenses found.</p>
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
								<Button variant="danger" onclick={() => handleDelete(exp.id)}>DEL</Button>
							</td>
						</tr>
					{/each}
				</tbody>
			</table>
		</div>

		<!-- Pagination -->
		<div class="flex items-center gap-2 mt-3 font-system text-sm">
			<Button onclick={goPrev} disabled={cursorStack.length === 0}>◀ PREV</Button>
			<Button onclick={goNext} disabled={!nextCursor}>NEXT ▶</Button>
		</div>
	{/if}
</Window>
