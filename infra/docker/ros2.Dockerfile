FROM ros:humble

# 必要なパッケージをインストール
RUN apt-get update && apt-get install -y --no-install-recommends \
    python3-pip \
    ros-humble-rosbridge-suite \
    && rm -rf /var/lib/apt/lists/*

# Python依存パッケージ
RUN pip3 install --no-cache-dir websockets pyyaml

# ワークスペース作成
WORKDIR /ros2_ws

# ソースコードをコピー
COPY ros2_ws/src /ros2_ws/src

# ビルド
RUN . /opt/ros/humble/setup.sh && \
    colcon build --symlink-install && \
    echo ". /ros2_ws/install/setup.bash" >> /root/.bashrc

# エントリポイント
COPY infra/docker/ros2_entrypoint.sh /ros2_entrypoint.sh
RUN chmod +x /ros2_entrypoint.sh

ENTRYPOINT ["/ros2_entrypoint.sh"]
CMD ["ros2", "launch", "robotasker_core", "robotasker.launch.py"]
