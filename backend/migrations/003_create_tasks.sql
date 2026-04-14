-- +migrate Up
CREATE TABLE tasks (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    robot_id UUID REFERENCES robots(id),
    original_text TEXT NOT NULL,
    parsed_action VARCHAR(50) NOT NULL,
    target_location_id UUID REFERENCES locations(id),
    priority INT NOT NULL DEFAULT 0,
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    result_message TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    started_at TIMESTAMP WITH TIME ZONE,
    completed_at TIMESTAMP WITH TIME ZONE
);

CREATE INDEX idx_tasks_status ON tasks(status);
CREATE INDEX idx_tasks_robot_id ON tasks(robot_id);

-- +migrate Down
DROP TABLE IF EXISTS tasks;
