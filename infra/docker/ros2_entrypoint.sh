#!/bin/bash
set -e

# ROS2環境を読み込む
source /opt/ros/humble/setup.bash
source /ros2_ws/install/setup.bash

exec "$@"
