"""
RoboTasker 全ノード一括起動用 launch ファイル

起動するノード:
  1. bridge_node       - GoバックエンドとのWebSocket通信
  2. task_executor_node - タスク受信・実行制御
  3. navigation_controller - 移動シミュレーション

使い方:
  ros2 launch robotasker_core robotasker.launch.py
  ros2 launch robotasker_core robotasker.launch.py robot_id:=robot-02 ws_port:=9091
"""

from launch import LaunchDescription
from launch.actions import DeclareLaunchArgument
from launch.substitutions import LaunchConfiguration
from launch_ros.actions import Node


def generate_launch_description():
    # Launch 引数
    robot_id_arg = DeclareLaunchArgument(
        'robot_id', default_value='robot-01',
        description='ロボット識別子')
    ws_port_arg = DeclareLaunchArgument(
        'ws_port', default_value='9090',
        description='WebSocketサーバーのポート番号')
    speed_arg = DeclareLaunchArgument(
        'speed', default_value='1.0',
        description='ロボットの移動速度 (m/s)')
    initial_x_arg = DeclareLaunchArgument(
        'initial_x', default_value='0.0',
        description='初期X座標')
    initial_y_arg = DeclareLaunchArgument(
        'initial_y', default_value='0.0',
        description='初期Y座標')

    # bridge_node
    bridge_node = Node(
        package='robotasker_core',
        executable='bridge_node',
        name='robotasker_bridge',
        output='screen',
        parameters=[{
            'ws_port': LaunchConfiguration('ws_port'),
        }],
    )

    # task_executor_node
    task_executor_node = Node(
        package='robotasker_core',
        executable='task_executor_node',
        name='task_executor',
        output='screen',
    )

    # navigation_controller
    navigation_controller = Node(
        package='robotasker_core',
        executable='navigation_controller',
        name='navigation_controller',
        output='screen',
        parameters=[{
            'robot_id': LaunchConfiguration('robot_id'),
            'speed': LaunchConfiguration('speed'),
            'initial_x': LaunchConfiguration('initial_x'),
            'initial_y': LaunchConfiguration('initial_y'),
        }],
    )

    return LaunchDescription([
        robot_id_arg,
        ws_port_arg,
        speed_arg,
        initial_x_arg,
        initial_y_arg,
        bridge_node,
        task_executor_node,
        navigation_controller,
    ])
