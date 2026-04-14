from setuptools import find_packages, setup
import os
from glob import glob

package_name = 'robotasker_core'

setup(
    name=package_name,
    version='0.1.0',
    packages=find_packages(exclude=['test']),
    data_files=[
        ('share/ament_index/resource_index/packages', ['resource/' + package_name]),
        ('share/' + package_name, ['package.xml']),
        (os.path.join('share', package_name, 'launch'), glob('launch/*.py')),
        (os.path.join('share', package_name, 'config'), glob('config/*.yaml')),
    ],
    install_requires=['setuptools', 'websockets'],
    zip_safe=True,
    maintainer='takaki',
    maintainer_email='takaki@example.com',
    description='RoboTasker コアノード群',
    license='MIT',
    entry_points={
        'console_scripts': [
            'bridge_node = robotasker_core.bridge_node:main',
            'task_executor_node = robotasker_core.task_executor_node:main',
            'navigation_controller = robotasker_core.navigation_controller:main',
        ],
    },
)
