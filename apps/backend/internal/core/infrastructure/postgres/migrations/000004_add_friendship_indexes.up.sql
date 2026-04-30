-- makes querying non-group expenses fast
CREATE INDEX IF NOT EXISTS idx_expenses_group_null ON expenses(group_id) WHERE group_id IS NULL;

-- speeds up the 'EXISTS' check to see if a user is involved in an expense
CREATE INDEX IF NOT EXISTS idx_splits_user_expense ON splits(user_id, expense_id);