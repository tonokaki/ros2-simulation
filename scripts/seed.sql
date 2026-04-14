-- ロボット初期データ
INSERT INTO robots (name, status, current_location, battery_level, position_x, position_y) VALUES
('RoboTasker-01', 'idle', '充電ステーション', 100, 0.0, 0.0),
('RoboTasker-02', 'idle', '充電ステーション', 95, 1.0, 0.0);

-- 移動可能地点
INSERT INTO locations (name, x, y, floor, location_type) VALUES
('充電ステーション', 0.0, 0.0, '1F', 'station'),
('受付',           5.0, 0.0, '1F', 'room'),
('会議室A',        3.0, 4.0, '1F', 'room'),
('会議室B',        7.0, 4.0, '1F', 'room'),
('休憩室',         5.0, 8.0, '1F', 'room'),
('倉庫',           0.0, 8.0, '1F', 'room'),
('エントランス',    10.0, 0.0, '1F', 'corridor');
