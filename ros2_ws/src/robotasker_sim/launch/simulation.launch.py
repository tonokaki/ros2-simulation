"""
RoboTasker Gazebo シミュレーション launch ファイル

起動内容:
  1. Gazebo サーバー + クライアント（オフィスワールド）
  2. TurtleBot3 スポーン（充電ステーション位置）
  3. RoboTasker コアノード群（bridge + executor + navigation）

前提:
  - TurtleBot3 パッケージがインストール済み
  - TURTLEBOT3_MODEL 環境変数が設定済み (burger 推奨)

使い方:
  export TURTLEBOT3_MODEL=burger
  ros2 launch robotasker_sim simulation.launch.py

  # ヘッドレス（GUI無し）:
  ros2 launch robotasker_sim simulation.launch.py gui:=false
"""

import os
from launch import LaunchDescription
from launch.actions import DeclareLaunchArgument, IncludeLaunchDescription, SetEnvironmentVariable
from launch.conditions import IfCondition
from launch.substitutions import LaunchConfiguration, PathJoinSubstitution
from launch_ros.actions import Node
from launch.launch_description_sources import PythonLaunchDescriptionSource
from ament_index_python.packages import get_package_share_directory


def generate_launch_description():
    # パッケージディレクトリ
    sim_pkg = get_package_share_directory('robotasker_sim')
    core_pkg = get_package_share_directory('robotasker_core')

    # Launch 引数
    gui_arg = DeclareLaunchArgument('gui', default_value='true', description='Gazebo GUIを起動するか')
    robot_id_arg = DeclareLaunchArgument('robot_id', default_value='robot-01', description='ロボットID')
    ws_port_arg = DeclareLaunchArgument('ws_port', default_value='9090', description='WebSocketポート')

    # ワールドファイルのパス
    world_path = os.path.join(sim_pkg, 'worlds', 'office.world')

    # TurtleBot3 モデル設定
    turtlebot3_model = SetEnvironmentVariable('TURTLEBOT3_MODEL', 'burger')

    # Gazebo サーバー起動
    gazebo_server = IncludeLaunchDescription(
        PythonLaunchDescriptionSource([
            PathJoinSubstitution([
                get_package_share_directory('gazebo_ros'),
                'launch', 'gzserver.launch.py'
            ])
        ]),
        launch_arguments={'world': world_path}.items(),
    )

    # Gazebo クライアント (GUI)
    gazebo_client = IncludeLaunchDescription(
        PythonLaunchDescriptionSource([
            PathJoinSubstitution([
                get_package_share_directory('gazebo_ros'),
                'launch', 'gzclient.launch.py'
            ])
        ]),
        condition=IfCondition(LaunchConfiguration('gui')),
    )

    # TurtleBot3 スポーン（充電ステーション位置: 0, 0）
    spawn_robot = Node(
        package='gazebo_ros',
        executable='spawn_entity.py',
        arguments=[
            '-entity', 'turtlebot3_burger',
            '-topic', 'robot_description',
            '-x', '0.0',
            '-y', '0.0',
            '-z', '0.01',
        ],
        output='screen',
    )

    # TurtleBot3 のロボットモデル(URDF)
    turtlebot3_state_publisher = IncludeLaunchDescription(
        PythonLaunchDescriptionSource([
            PathJoinSubstitution([
                get_package_share_directory('turtlebot3_gazebo'),
                'launch', 'robot_state_publisher.launch.py'
            ])
        ]),
        launch_arguments={'use_sim_time': 'true'}.items(),
    )

    # RoboTasker コアノード群
    robotasker_core = IncludeLaunchDescription(
        PythonLaunchDescriptionSource([
            os.path.join(core_pkg, 'launch', 'robotasker.launch.py')
        ]),
        launch_arguments={
            'robot_id': LaunchConfiguration('robot_id'),
            'ws_port': LaunchConfiguration('ws_port'),
        }.items(),
    )

    return LaunchDescription([
        gui_arg,
        robot_id_arg,
        ws_port_arg,
        turtlebot3_model,
        gazebo_server,
        gazebo_client,
        spawn_robot,
        turtlebot3_state_publisher,
        robotasker_core,
    ])
