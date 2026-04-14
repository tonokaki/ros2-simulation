# バックエンドからロボットへのタスク指令
string task_id
string robot_id
string action          # deliver, patrol, return, goto
string target_location # 目的地名
float64 target_x       # 目的地X座標
float64 target_y       # 目的地Y座標
int32 priority         # 優先度 (0=normal, 10=high)
