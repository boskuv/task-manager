-- +goose Up
CREATE INDEX idx_tasks_team_id_status ON tasks (team_id, status);

-- assignee_id is already indexed by fk_tasks_assignee_id (created with foreign key).

CREATE INDEX idx_team_members_user_id_team_id ON team_members (user_id, team_id);

CREATE INDEX idx_task_history_task_id_created_at ON task_history (task_id, created_at);

-- +goose Down
DROP INDEX idx_task_history_task_id_created_at ON task_history;
DROP INDEX idx_team_members_user_id_team_id ON team_members;
DROP INDEX idx_tasks_team_id_status ON tasks;
