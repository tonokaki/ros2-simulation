"""
bridge_node: GoバックエンドとROS2の橋渡しノード

GoバックエンドからWebSocket経由でタスク指令を受信し、ROS2トピックに変換する。
逆にROS2のタスク状態・ロボット状態をGoバックエンドにWebSocket経由で送信する。

通信プロトコル（バックエンド ↔ bridge_node）:
  バックエンド → bridge: {"type": "task_command", "data": {...}}
  bridge → バックエンド: {"type": "task_status", "data": {...}}
  bridge → バックエンド: {"type": "robot_state", "data": {...}}
"""

import json
import asyncio
import threading

import rclpy
from rclpy.node import Node

from robotasker_msgs.msg import TaskCommand, TaskStatus, RobotState

try:
    import websockets
    from websockets.asyncio.server import serve as ws_serve
    HAS_WEBSOCKETS = True
except ImportError:
    HAS_WEBSOCKETS = False


class BridgeNode(Node):
    """GoバックエンドとROS2を接続するブリッジノード"""

    def __init__(self):
        super().__init__('robotasker_bridge')

        # パラメータ
        self.declare_parameter('ws_port', 9090)
        self.ws_port = self.get_parameter('ws_port').value

        # Publisher: バックエンドからのタスク指令をROS2に送信
        self.task_command_pub = self.create_publisher(TaskCommand, '/task/command', 10)

        # Subscriber: ROS2からのタスク状態・ロボット状態をバックエンドに送信
        self.task_status_sub = self.create_subscription(
            TaskStatus, '/task/status', self.on_task_status, 10)
        self.robot_state_sub = self.create_subscription(
            RobotState, '/robot/state', self.on_robot_state, 10)

        # ロボットIDマッピング: ローカルID → バックエンドUUID
        # task_command受信時にUUIDを学習し、robot_state送信時に変換する
        self.robot_id_map: dict = {}  # local_id -> backend_uuid

        # 接続中のWebSocketクライアント
        self.ws_clients: set = set()
        self.loop = None

        # WebSocketサーバーをバックグラウンドスレッドで起動
        if HAS_WEBSOCKETS:
            self.ws_thread = threading.Thread(target=self._run_ws_server, daemon=True)
            self.ws_thread.start()
            self.get_logger().info(f'WebSocketサーバー起動: ws://0.0.0.0:{self.ws_port}')
        else:
            self.get_logger().warn('websocketsパッケージがありません。WebSocketサーバーは無効です。')

        # パラメータ: ローカルロボットID（navigation_controllerと合わせる）
        self.declare_parameter('local_robot_id', 'robot-01')
        self.local_robot_id = self.get_parameter('local_robot_id').value

        self.get_logger().info('BridgeNode 起動完了')

    def _run_ws_server(self):
        """WebSocketサーバーを別スレッドの asyncio イベントループで動かす"""
        self.loop = asyncio.new_event_loop()
        asyncio.set_event_loop(self.loop)
        self.loop.run_until_complete(self._ws_server_main())

    async def _ws_server_main(self):
        async with ws_serve(self._ws_handler, '0.0.0.0', self.ws_port):
            await asyncio.Future()  # 永久に待機

    async def _ws_handler(self, websocket):
        """WebSocket接続ハンドラ"""
        self.ws_clients.add(websocket)
        client_addr = websocket.remote_address
        self.get_logger().info(f'WebSocket接続: {client_addr}')
        try:
            async for raw_message in websocket:
                await self._handle_ws_message(raw_message)
        except websockets.exceptions.ConnectionClosed:
            pass
        finally:
            self.ws_clients.discard(websocket)
            self.get_logger().info(f'WebSocket切断: {client_addr}')

    async def _handle_ws_message(self, raw: str):
        """バックエンドからのWebSocketメッセージを処理する"""
        try:
            msg = json.loads(raw)
        except json.JSONDecodeError:
            self.get_logger().warn(f'不正なJSON: {raw[:100]}')
            return

        msg_type = msg.get('type', '')
        data = msg.get('data', {})

        if msg_type == 'task_command':
            # バックエンドのUUID→ローカルIDにマッピング
            backend_robot_id = data.get('robot_id', '')
            self.robot_id_map[self.local_robot_id] = backend_robot_id

            # タスク指令をROS2トピックに変換（robot_idはローカルIDに変換）
            cmd = TaskCommand()
            cmd.task_id = data.get('task_id', '')
            cmd.robot_id = self.local_robot_id
            cmd.action = data.get('action', '')
            cmd.target_location = data.get('target_location', '')
            cmd.target_x = float(data.get('target_x', 0.0))
            cmd.target_y = float(data.get('target_y', 0.0))
            cmd.priority = int(data.get('priority', 0))
            self.task_command_pub.publish(cmd)
            self.get_logger().info(
                f'タスク指令受信: task={cmd.task_id[:8]} action={cmd.action} target={cmd.target_location} '
                f'robot={self.local_robot_id} (backend={backend_robot_id[:8]})')
        else:
            self.get_logger().warn(f'不明なメッセージタイプ: {msg_type}')

    def _resolve_robot_id(self, local_id: str) -> str:
        """ローカルIDをバックエンドUUIDに変換する"""
        return self.robot_id_map.get(local_id, local_id)

    def on_task_status(self, msg: TaskStatus):
        """ROS2のタスク状態をWebSocket経由でバックエンドに送信"""
        backend_id = self._resolve_robot_id(msg.robot_id)
        payload = json.dumps({
            'type': 'task_status',
            'data': {
                'task_id': msg.task_id,
                'robot_id': backend_id,
                'status': msg.status,
                'message': msg.message,
            }
        })
        self._broadcast_ws(payload)
        self.get_logger().info(f'タスク状態送信: task={msg.task_id[:8]} status={msg.status}')

    def on_robot_state(self, msg: RobotState):
        """ROS2のロボット状態をWebSocket経由でバックエンドに送信"""
        backend_id = self._resolve_robot_id(msg.robot_id)
        payload = json.dumps({
            'type': 'robot_state',
            'data': {
                'robot_id': backend_id,
                'status': msg.status,
                'position_x': msg.position_x,
                'position_y': msg.position_y,
                'velocity': msg.velocity,
                'current_location': msg.current_location,
                'battery_level': msg.battery_level,
            }
        })
        self._broadcast_ws(payload)

    def _broadcast_ws(self, payload: str):
        """全WebSocketクライアントにメッセージを送信"""
        if not self.loop or not self.ws_clients:
            return
        asyncio.run_coroutine_threadsafe(
            self._async_broadcast(payload), self.loop
        )

    async def _async_broadcast(self, payload: str):
        disconnected = set()
        for ws in self.ws_clients:
            try:
                await ws.send(payload)
            except Exception:
                disconnected.add(ws)
        self.ws_clients -= disconnected


def main(args=None):
    rclpy.init(args=args)
    node = BridgeNode()
    try:
        rclpy.spin(node)
    except KeyboardInterrupt:
        pass
    finally:
        node.destroy_node()
        rclpy.shutdown()


if __name__ == '__main__':
    main()
