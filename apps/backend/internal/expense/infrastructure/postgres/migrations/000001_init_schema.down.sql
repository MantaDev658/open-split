-- drop in reverse order of creation to avoid foreign key constraint errors
DROP TABLE IF EXISTS splits;
DROP TABLE IF EXISTS expenses;
DROP TABLE IF EXISTS users;