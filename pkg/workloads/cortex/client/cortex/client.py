# Copyright 2019 Cortex Labs, Inc.
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

import pathlib
from pathlib import Path
import os
import types
import subprocess
import sys
import shutil
import yaml
import urllib.parse
import base64
import inspect

import dill
import requests
from requests.exceptions import HTTPError
import msgpack


class Client(object):
    def __init__(self, aws_access_key_id, aws_secret_access_key, operator_url):
        """Initialize a Client to a Cortex Operator

        Args:
            aws_access_key_id (string): AWS access key associated with the account that the cluster is running on
            aws_secret_access_key (string): AWS secret key associated with the AWS access key
            operator_url (string): operator URL of your cluster
        """

        self.operator_url = operator_url
        self.workspace = str(Path.home() / ".cortex" / "workspace")
        self.aws_access_key_id = aws_access_key_id
        self.aws_secret_access_key = aws_secret_access_key
        self.headers = {
            "CortexAPIVersion": "master",  # CORTEX_VERSION
            "Authorization": "CortexAWS {}|{}".format(
                self.aws_access_key_id, self.aws_secret_access_key
            ),
        }

        pathlib.Path(self.workspace).mkdir(parents=True, exist_ok=True)

    def deploy(
        self, deployment_name, api_name, predictor, model_path=None, tf_serving_key=None, config={}
    ):
        """Deploy an API

        Args:
            deployment_name (string): deployment name
            api_name (string): API name
            predictor (class): class definition implementing the Cortex Predictor interface based on your model format
            model_path (string): S3 path to model (required for TensorFlowPredictor and ONNXPredictor)
            tf_serving_key (string, optional): name of the signature def to use for prediction (required for TensorFlowPredictor if your model has more than one signature def)
            config (dict): dictionary passed to the constructor of a Predictor

        Returns:
            string: url to the deployed API
        """

        working_dir = os.path.join(self.workspace, deployment_name)
        api_working_dir = os.path.join(working_dir, api_name)
        pathlib.Path(api_working_dir).mkdir(parents=True, exist_ok=True)

        api_config = {"kind": "api", "name": api_name}

        if not inspect.isclass(predictor):
            raise Exception(
                "predictor should be a class definition implementing one of the following Cortex Predictor interfaces: PythonPredictor, TensorFlowPredictor, ONNXPredictor"
            )

        class_name = predictor.__name__

        if (
            class_name != "PythonPredictor"
            and class_name != "TensorFlowPredictor"
            and class_name != "ONNXPredictor"
        ):
            raise Exception(
                "unexpected class name found: expected PythonPredictor, TensorFlowPredictor or ONNXPredictor but found "
                + class_name
            )

        if class_name == "PythonPredictor":
            model_format = "python"
        elif class_name == "TensorFlowPredictor":
            model_format = "tensorflow"
        else:
            model_format = "onnx"

        api_config[model_format] = {}

        if model_path is not None:
            api_config[model_format]["model"] = model_path

        if model_format == "tensorflow" and tf_serving_key is not None:
            api_config[model_format]["serving_key"] = tf_serving_key

        reqs = subprocess.check_output([sys.executable, "-m", "pip", "freeze"])

        with open(os.path.join(api_working_dir, "requirements.txt"), "w") as f:
            f.writelines(reqs.decode())

        with open(os.path.join(api_working_dir, "predictor.pickle"), "wb") as f:
            dill.dump(predictor, f, recurse=True)

        api_config[model_format]["predictor"] = "predictor.pickle"

        deployment_config = [{"kind": "deployment", "name": deployment_name}, api_config]

        cortex_yaml_path = os.path.join(working_dir, "cortex.yaml")
        with open(cortex_yaml_path, "w") as f:
            f.write(yaml.dump(deployment_config))

        project_zip_path = os.path.join(working_dir, "project")
        shutil.make_archive(project_zip_path, "zip", api_working_dir)
        project_zip_path += ".zip"

        queries = {"force": "false", "ignoreCache": "false"}

        with open(cortex_yaml_path, "rb") as config, open(project_zip_path, "rb") as project:
            files = {"cortex.yaml": config, "project.zip": project}
            try:
                resp = requests.post(
                    urllib.parse.urljoin(self.operator_url, "deploy"),
                    params=queries,
                    files=files,
                    headers=self.headers,
                    verify=False,
                )
                resp.raise_for_status()
                resources = resp.json()
            except HTTPError as err:
                resp = err.response
                if "error" in resp.json():
                    raise Exception(resp.json()["error"]) from err
                raise

            b64_encoded_context = resources["context"]
            context_msgpack_bytestring = base64.b64decode(b64_encoded_context)
            ctx = msgpack.loads(context_msgpack_bytestring, raw=False)
            return urllib.parse.urljoin(resources["apis_base_url"], ctx["apis"][api_name]["path"])
