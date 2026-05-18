-- Down migration avoids dropping tables if it means losing data. 
-- For SQLite, removing FKs means dropping and recreating tables again.
-- In development, it is often simpler to just rollback prior schemas fully.
-- Here we'll just leave the tables as is because dropping constraints in SQLite is identical to recreating them.

-- No-op for SQLite rollback of FK additions, since the columns themselves are correct.
