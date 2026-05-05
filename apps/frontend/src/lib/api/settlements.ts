import { apiFetch } from './client';

export interface CreateSettlementInput {
	receiver_id: string;
	amount_cents: number;
	group_id?: string;
}

export function createSettlement(input: CreateSettlementInput) {
	return apiFetch<void>('/settlements', { method: 'POST', body: JSON.stringify(input) });
}
