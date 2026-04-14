from setuptools import find_packages, setup
import os
from glob import glob

package_name = 'robotasker_sim'

setup(
    name=package_name,
    version='0.1.0',
    packages=find_packages(exclude=['test']),
    data_files=[
        ('share/ament_index/resource_index/packages', ['resource/' + package_name]),
        ('share/' + package_name, ['package.xml']),
        (os.path.join('share', package_name, 'launch'), glob('launch/*.py')),
        (os.path.join('share', package_name, 'worlds'), glob('worlds/*.world')),
    ],
    install_requires=['setuptools'],
    zip_safe=True,
    maintainer='takaki',
    maintainer_email='takaki@example.com',
    description='RoboTasker Gazebo シミュレーション環境',
    license='MIT',
    entry_points={},
)
