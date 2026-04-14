"""
navigation_controller: 簡易移動シミュレーションノード

/navigation/goal を受信し、ロボットをその座標に向けて直線移動させる。
移動中は定期的に /robot/state を発行し、到着時に current_location を設定する。

Phase 4 で Nav2 に差し替え可能な設計:
  - /navigation/goal (PoseStamped) を受信するインターフェースは同じ
  - /cmd_vel (Twist) の発行は Gazebo 連携時にも使える
  - /robot/state は独自だが、Nav2 + tf2 からの変換も可能

簡易シミュレーション仕様:
  - 移動速度: 1.0 m/s (パラメータで変更可)
  - 更新周期: 10Hz
  - 到着判定: 目的地から 0.3m 以内
"""

import math
import yaml

import rclpy
from rclpy.node import Node
from ament_index_python.packages import get_package_share_directory

from robotasker_msgs.msg import RobotState
from geometry_msgs.msg import PoseStamped, Twist


class NavigationController(Node):
    """簡易ナビゲーション制御ノード（直線移動シミュレーション）"""

    def __init__(self):
        super().__init__('navigation_controller')

        # パラメータ
        self.declare_parameter('robot_id', 'robot-01')
        self.declare_parameter('speed', 1.0)          # m/s
        self.declare_parameter('update_rate', 10.0)    # Hz
        self.declare_parameter('arrival_threshold', 0.3)  # m
        self.declare_parameter('initial_x', 0.0)
        self.declare_parameter('initial_y', 0.0)

        self.robot_id = self.get_parameter('robot_id').value
        self.speed = self.get_parameter('speed').value
        self.update_rate = self.get_parameter('update_rate').value
        self.arrival_threshold = self.get_parameter('arrival_threshold').value

        # ロボット状態
        self.pos_x = self.get_parameter('initial_x').value
        self.pos_y = self.get_parameter('initial_y').value
        self.battery = 100
        self.status = 'idle'

        # 目標座標
        self.goal_x = None
        self.goal_y = None
        self.goal_location_name = ''

        # 地点名のリバースルックアップ用
        self.locations = self._load_locations()

        # Subscriber
        self.goal_sub = self.create_subscription(
            PoseStamped, '/navigation/goal', self.on_goal, 10)

        # Publisher
        self.state_pub = self.create_publisher(RobotState, '/robot/state', 10)
        self.cmd_vel_pub = self.create_publisher(Twist, '/cmd_vel', 10)

        # 定期更新タイマー
        period = 1.0 / self.update_rate
        self.timer = self.create_timer(period, self.update)

        self.get_logger().info(
            f'NavigationController 起動: robot_id={self.robot_id} '
            f'pos=({self.pos_x:.1f}, {self.pos_y:.1f}) speed={self.speed}m/s')

    def _load_locations(self) -> dict:
        """地点座標をYAMLから読み込む"""
        try:
            pkg_dir = get_package_share_directory('robotasker_core')
            config_path = f'{pkg_dir}/config/locations.yaml'
            with open(config_path, 'r') as f:
                data = yaml.safe_load(f)
            locations = {}
            for name, info in data.get('locations', {}).items():
                locations[name] = (info['x'], info['y'])
            self.get_logger().info(f'地点データ読み込み: {len(locations)}件')
            return locations
        except Exception as e:
            self.get_logger().warn(f'地点データ読み込み失敗: {e}（ハードコード値を使用）')
            return {
                '充電ステーション': (0.0, 0.0),
                '受付': (5.0, 0.0),
                '会議室A': (3.0, 4.0),
                '会議室B': (7.0, 4.0),
                '休憩室': (5.0, 8.0),
                '倉庫': (0.0, 8.0),
                'エントランス': (10.0, 0.0),
            }

    def _find_location_name(self, x: float, y: float) -> str:
        """座標から最も近い地点名を返す（到着判定用）"""
        for name, (lx, ly) in self.locations.items():
            dist = math.sqrt((x - lx) ** 2 + (y - ly) ** 2)
            if dist < self.arrival_threshold:
                return name
        return ''

    def on_goal(self, msg: PoseStamped):
        """ナビゲーション目標を受信"""
        self.goal_x = msg.pose.position.x
        self.goal_y = msg.pose.position.y
        self.goal_location_name = self._find_location_name(self.goal_x, self.goal_y)
        self.status = 'moving'
        self.get_logger().info(
            f'目標受信: ({self.goal_x:.1f}, {self.goal_y:.1f}) → {self.goal_location_name or "不明"}')

    def update(self):
        """定期更新: 移動シミュレーション + 状態発行"""
        velocity = 0.0

        if self.status == 'moving' and self.goal_x is not None:
            # 目的地までの距離と方向
            dx = self.goal_x - self.pos_x
            dy = self.goal_y - self.pos_y
            dist = math.sqrt(dx * dx + dy * dy)

            if dist < self.arrival_threshold:
                # 到着
                self.pos_x = self.goal_x
                self.pos_y = self.goal_y
                self.status = 'idle'
                velocity = 0.0

                # cmd_vel 停止
                twist = Twist()
                self.cmd_vel_pub.publish(twist)

                self.get_logger().info(
                    f'到着: ({self.pos_x:.1f}, {self.pos_y:.1f}) = {self.goal_location_name}')
            else:
                # 移動
                dt = 1.0 / self.update_rate
                move_dist = min(self.speed * dt, dist)
                ratio = move_dist / dist
                self.pos_x += dx * ratio
                self.pos_y += dy * ratio
                velocity = self.speed

                # cmd_vel 発行（Gazebo連携用）
                twist = Twist()
                angle = math.atan2(dy, dx)
                twist.linear.x = self.speed * math.cos(angle)
                twist.linear.y = self.speed * math.sin(angle)
                self.cmd_vel_pub.publish(twist)

                # バッテリー消費（移動距離に応じて微減）
                self.battery = max(0, self.battery - int(move_dist * 0.1))

        # ロボット状態発行
        state = RobotState()
        state.robot_id = self.robot_id
        state.status = self.status
        state.position_x = round(self.pos_x, 2)
        state.position_y = round(self.pos_y, 2)
        state.velocity = round(velocity, 2)
        state.battery_level = self.battery

        # 到着時のみ地点名を設定
        if self.status == 'idle' and self.goal_location_name:
            state.current_location = self.goal_location_name
        else:
            state.current_location = ''

        self.state_pub.publish(state)


def main(args=None):
    rclpy.init(args=args)
    node = NavigationController()
    try:
        rclpy.spin(node)
    except KeyboardInterrupt:
        pass
    finally:
        node.destroy_node()
        rclpy.shutdown()


if __name__ == '__main__':
    main()
