#!/bin/bash

# Copyright 2020 Cortex Labs, Inc.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

set -e

export PYTHONPATH=$PYTHONPATH:$PYTHON_PATH

if [ -f "/mnt/project/requirements.txt" ]; then
    pip --no-cache-dir install -r /mnt/project/requirements.txt
fi

mkdir -p /mnt/project

cd /mnt/project

echo $DOWNLOAD_CONFIG

/usr/bin/python3.6 /src/cortex/serve/download.py --download=ewogICJkb3dubG9hZF9hcmdzIjogWwogICAgewogICAgICAiZnJvbSI6ICJzMzovL2NvcnRleC1jbHVzdGVyLXZpc2hhbC9wcm9qZWN0cy8yOGNjZDZmOTM4YTM4MmRlNTcxMjcwYTMyMjRjMmQ3ZGE3MWM2ZGI5ZTc0Nzg0NTVkNmM0ODY0ZmI0MmJmNzIuemlwIiwKICAgICAgInRvIjogIi9tbnQvcHJvamVjdCIsCiAgICAgICJ1bnppcCI6IHRydWUsCiAgICAgICJpdGVtX25hbWUiOiAidGhlIHByb2plY3QgY29kZSIsCiAgICAgICJ0Zl9tb2RlbF92ZXJzaW9uX3JlbmFtZSI6ICIiLAogICAgICAiaGlkZV9mcm9tX2xvZyI6IHRydWUsCiAgICAgICJoaWRlX3VuemlwcGluZ19sb2ciOiB0cnVlCiAgICB9CiAgXSwKICAibGFzdF9sb2ciOiAicHVsbGluZyB0aGUgcHl0aG9uIHNlcnZpbmcgaW1hZ2UiCn0=

export PORT="${MY_PORT:-8888}"

export MY_PORT="8888"

gunicorn -b 0.0.0.0:$PORT --access-logfile=- --pythonpath=$PYTHONPATH --chdir /mnt/project --log-level debug cortex.serve.wsgi:app
