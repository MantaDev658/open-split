import { apiFetch } from './client';
import type { AuditLog, CreateGroupResponse, Group, Paginated } from './types';

export function listGroups() {
	return apiFetch<Group[]>('/groups');
}

export function createGroup(name: string) {
	return apiFetch<CreateGroupResponse>('/groups', { method: 'POST', body: JSON.stringify({ name }) });
}

export function updateGroup(id: string, name: string) {
	return apiFetch<void>(`/groups/${id}`, { method: 'PUT', body: JSON.stringify({ name }) });
}

export function deleteGroup(id: string) {
	return apiFetch<void>(`/groups/${id}`, { method: 'DELETE' });
}

export function addGroupMember(groupID: string, userID: string) {
	return apiFetch<{ status: string }>(`/groups/${groupID}/members`, {
		method: 'POST',
		body: JSON.stringify({ user_id: userID })
	});
}

export function removeGroupMember(groupID: string, userID: string) {
	return apiFetch<void>(`/groups/${groupID}/members/${userID}`, { method: 'DELETE' });
}

export function getGroupActivity(groupID: string, cursor?: string, limit = 20) {
	const params = new URLSearchParams();
	if (cursor) params.set('cursor', cursor);
	params.set('limit', String(limit));
	return apiFetch<Paginated<AuditLog>>(`/groups/${groupID}/activity?${params}`);
}
