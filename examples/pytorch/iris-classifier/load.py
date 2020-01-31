import asyncio
import aiohttp
import pandas as pd
import json

data = None
with open("sample.json", "r") as f:
    data = json.load(f)

datas = [data] * 100000  # tune number of requests
url = "http://af588dc54421b11ea9a9c06e4cdecb72-509053161.us-west-2.elb.amazonaws.com:80/predict"


async def process(url: str, datas, timeout=3600, n_conns=40):
    async def post(data, i):
        print(i)
        try:
            async with session.post(
                url, json=data, headers={"Host": "yolov4.default.example.com"}
            ) as response:
                print(response)
                return await response.json()
        except Exception as ex:
            return ex

    async with aiohttp.ClientSession(
        connector=aiohttp.TCPConnector(limit=n_conns), timeout=aiohttp.ClientTimeout(total=timeout)
    ) as session:

        tasks = [asyncio.ensure_future(post(data, i)) for i, data in enumerate(datas)]
        responses = await asyncio.gather(*tasks)
        flat_resp = responses


loop = asyncio.get_event_loop()
task = asyncio.ensure_future(process(url, datas))
loop.run_until_complete(task)
