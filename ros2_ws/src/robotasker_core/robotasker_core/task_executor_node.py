"""
task_executor_node: タスクの実行管理ノード

/task/command を受信し、アクションに応じてナビゲーション指令を発行する。
ナビゲーション完了を検知してタスク完了ステータスを発行する。

状態遷移:
  IDLE → EXECUTING (タスク受信)
  EXECUTING → IDLE (ナビゲーション完了/失敗)
"""

import rclpy
from rclpy.node import Node

from robotasker_msgs.msg import TaskCommand, TaskStatus, RobotState
from geometry_msgs.msg import PoseStamped


class TaskExecutorNode(Node):
    """タスクの受信と実行管理を行うノード"""

    def __init__(self):
        super().__init__('task_executor')

        # 現在実行中のタスク
        self.current_task: TaskCommand | None = None

        # Subscriber: タスク指令
        self.task_command_sub = self.create_subscription(
            TaskCommand, '/task/command', self.on_task_command, 10)

        # Subscriber: ロボット状態（ナビゲーション完了検知用）
        self.robot_state_sub = self.create_subscription(
            RobotState, '/robot/state', self.on_robot_state, 10)

        # Publisher: タスク状態
        self.task_status_pub = self.create_publisher(TaskStatus, '/task/status', 10)

        # Publisher: ナビゲーション目標
        self.nav_goal_pub = self.create_publisher(PoseStamped, '/navigation/goal', 10)

        # タイムアウト用タイマー (60秒)
        self.timeout_timer = None
        self.TASK_TIMEOUT_SEC = 60.0

        self.get_logger().info('TaskExecutorNode 起動完了')

    def on_task_command(self, msg: TaskCommand):
        """タスク指令を受信してナビゲーション目標を発行する"""
        if self.current_task is not None:
            self.get_logger().warn(
                f'タスク実行中のため拒否: current={self.current_task.task_id[:8]} new={msg.task_id[:8]}')
            # 拒否ステータスを返す
            status = TaskStatus()
            status.task_id = msg.task_id
            status.robot_id = msg.robot_id
            status.status = 'failed'
            status.message = 'ロボットは別のタスクを実行中です'
            self.task_status_pub.publish(status)
            return

        self.current_task = msg
        self.get_logger().info(
            f'タスク開始: task={msg.task_id[:8]} action={msg.action} '
            f'target={msg.target_location} ({msg.target_x:.1f}, {msg.target_y:.1f})')

        # ステータス: in_progress
        status = TaskStatus()
        status.task_id = msg.task_id
        status.robot_id = msg.robot_id
        status.status = 'in_progress'
        status.message = f'{msg.target_location}へ移動中'
        self.task_status_pub.publish(status)

        # ナビゲーション目標を発行
        goal = PoseStamped()
        goal.header.stamp = self.get_clock().now().to_msg()
        goal.header.frame_id = 'map'
        goal.pose.position.x = msg.target_x
        goal.pose.position.y = msg.target_y
        goal.pose.position.z = 0.0
        goal.pose.orientation.w = 1.0
        self.nav_goal_pub.publish(goal)

        # タイムアウトタイマー開始
        if self.timeout_timer:
            self.timeout_timer.cancel()
        self.timeout_timer = self.create_timer(
            self.TASK_TIMEOUT_SEC, self.on_timeout)

    def on_robot_state(self, msg: RobotState):
        """ロボット状態を監視し、目的地到着を検知する"""
        if self.current_task is None:
            return
        if msg.robot_id != self.current_task.robot_id:
            return

        # 目的地到着判定: current_location が設定されていたら到着とみなす
        if msg.current_location and msg.current_location == self.current_task.target_location:
            self.get_logger().info(
                f'目的地到着: task={self.current_task.task_id[:8]} location={msg.current_location}')
            self._complete_task(True, f'{msg.current_location}に到着しました')

    def on_timeout(self):
        """タスクタイムアウト処理"""
        if self.current_task:
            self.get_logger().warn(f'タスクタイムアウト: task={self.current_task.task_id[:8]}')
            self._complete_task(False, 'タイムアウト: 指定時間内に目的地に到着できませんでした')

    def _complete_task(self, success: bool, message: str):
        """タスクを完了/失敗としてステータスを発行する"""
        if self.current_task is None:
            return

        status = TaskStatus()
        status.task_id = self.current_task.task_id
        status.robot_id = self.current_task.robot_id
        status.status = 'completed' if success else 'failed'
        status.message = message
        self.task_status_pub.publish(status)

        self.get_logger().info(
            f'タスク{"完了" if success else "失敗"}: task={self.current_task.task_id[:8]} - {message}')

        # リセット
        self.current_task = None
        if self.timeout_timer:
            self.timeout_timer.cancel()
            self.timeout_timer = None


def main(args=None):
    rclpy.init(args=args)
    node = TaskExecutorNode()
    try:
        rclpy.spin(node)
    except KeyboardInterrupt:
        pass
    finally:
        node.destroy_node()
        rclpy.shutdown()


if __name__ == '__main__':
    main()
