-- +goose Up
-- SQL section 'Up' is executed when this migration is applied

CREATE TABLE IF NOT EXISTS users (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,      
    email TEXT NOT NULL UNIQUE,
    password TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS daily_tasks (
    id TEXT PRIMARY KEY,
    title TEXT NOT NULL,
    is_completed BOOLEAN DEFAULT FALSE,
    date DATE NOT NULL,
    user_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE
);
CREATE INDEX IF NOT EXISTS idx_tasks_date ON daily_tasks(date);

CREATE TABLE IF NOT EXISTS diary_entries (
    id TEXT PRIMARY KEY,
    text TEXT,
    date DATE NOT NULL,
    user_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE
);
CREATE INDEX IF NOT EXISTS idx_diary_date ON diary_entries(date);

CREATE TABLE IF NOT EXISTS goal_groups (
    id TEXT PRIMARY KEY,
    title TEXT NOT NULL,
    user_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS yearly_goals (
    id TEXT PRIMARY KEY,
    title TEXT NOT NULL,
    total_steps INT DEFAULT 0,
    current_step INT DEFAULT 0,
    goal_group_id TEXT REFERENCES goal_groups(id) ON DELETE CASCADE
);
CREATE INDEX IF NOT EXISTS idx_goals_group ON yearly_goals(goal_group_id);

-- +goose Down
-- SQL section 'Down' is executed when this migration is rolled back

DROP TABLE IF EXISTS yearly_goals;
DROP TABLE IF EXISTS goal_groups;
DROP TABLE IF EXISTS diary_entries;
DROP TABLE IF EXISTS daily_tasks;
DROP TABLE IF EXISTS users;