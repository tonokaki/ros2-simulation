# ロボットの現在状態（定期的に発行）
string robot_id
string status          # idle, moving, executing, charging, error
float64 position_x
float64 position_y
float64 velocity       # 現在速度
string current_location # 到着した地点名（移動中は空文字）
int32 battery_level    # バッテリー残量 (0-100)
