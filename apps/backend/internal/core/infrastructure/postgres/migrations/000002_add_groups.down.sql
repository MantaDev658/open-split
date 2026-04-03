-- reverse order of creation
ALTER TABLE expenses DROP COLUMN IF EXISTS group_id;
DROP TABLE IF EXISTS group_members;
DROP TABLE IF EXISTS groups;