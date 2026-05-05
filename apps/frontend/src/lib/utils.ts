/** Formats an integer cent value as a dollar string, e.g. 5050 → "$50.50" */
export function formatCents(cents: number): string {
	return new Intl.NumberFormat('en-US', { style: 'currency', currency: 'USD' }).format(
		Math.abs(cents) / 100
	);
}

/** Formats an ISO date string as a short locale date, e.g. "Jan 5, 2026" */
export function formatDate(iso: string): string {
	return new Intl.DateTimeFormat('en-US', { month: 'short', day: 'numeric', year: 'numeric' }).format(
		new Date(iso)
	);
}
