#!/usr/bin/env -S /Users/Jubicudis/Tranquility-Neuro-OS/systems/python/venv311/bin/python3.11

import asyncio

import websockets


async def test():
    try:
        ws = await websockets.connect('ws://localhost:9001')
    except Exception as e:
        print(f'[DIAG][test_ws_connect] Failed to connect to ws://localhost:9001: {e}')
        raise
    print('WebSocket connection: OK')
    await ws.close()

if __name__ == "__main__":
    asyncio.run(test())
