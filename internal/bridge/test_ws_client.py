#!/usr/bin/env -S /Users/Jubicudis/Tranquility-Neuro-OS/systems/python/venv311/bin/python3.11

import asyncio
import hashlib
import json
import time

import websockets


def pretty_print(label, data):
    print(f"\n=== {label} ===")
    if isinstance(data, str):
        try:
            data = json.loads(data)
        except Exception:
            pass
    print(json.dumps(data, indent=2) if isinstance(data, dict) else data)


async def test_ws():
    try:
        try:
            ws = await websockets.connect('ws://localhost:9001')
        except Exception as e:
            print(f'[DIAG][test_ws_client] Failed to connect to ws://localhost:9001: {e}')
            raise
        print('WebSocket connection: OK')
        # QHP handshake (init phase)
        handshake_init = {
            "type": "qhp_handshake",
            "content": {"phase": "init"},
            "context": {
                "who": "TestClient",
                "what": "QHPHandshake",
                "when": time.time(),
                "where": "test_ws_client.py",
                "why": "IntegrationTest",
                "how": "WebSocket",
                "extent": 1.0
            }
        }
        await ws.send(json.dumps(handshake_init))
        challenge_msg = await ws.recv()
        pretty_print("QHP Challenge", challenge_msg)
        challenge_data = json.loads(challenge_msg)
        challenge = challenge_data.get("challenge")
        # QHP handshake (response phase)
        handshake_response = {
            "type": "qhp_handshake_ack",
            "challenge_response": hashlib.sha256((challenge + handshake_init["context"]["who"]).encode("utf-8")).hexdigest() if challenge else ""
        }
        await ws.send(json.dumps(handshake_response))
        # Wait for handshake completion or error
        result_msg = await ws.recv()
        pretty_print("QHP Result", result_msg)
        # 1. Basic ping (now after handshake)
        ping_msg = {
            "operation": "ping",
            "context": handshake_init["context"]
        }
        await ws.send(json.dumps(ping_msg))
        response = await ws.recv()
        pretty_print("Ping Response", response)
        # 2. Compression request (MCP style)
        compression_msg = {
            "operation": "compression",
            "data": {
                "input": "The quick brown fox jumps over the lazy dog. The quick brown fox jumps over the lazy dog. The quick brown fox jumps over the lazy dog.",
                "algorithm": "mobius_collapse_v3"
            },
            "context": handshake_init["context"]
        }
        await ws.send(json.dumps(compression_msg))
        comp_response = await ws.recv()
        pretty_print("Compression Response", comp_response)
        # Print compression ratio if present
        try:
            comp_data = json.loads(comp_response)
            if "data" in comp_data and isinstance(comp_data["data"], dict):
                ratio = comp_data["data"].get("compression_ratio")
                if ratio is not None:
                    print(f"Compression Ratio: {ratio}")
        except Exception:
            pass
        # 3. Context message
        context_msg = {
            "operation": "context",
            "context": handshake_init["context"]
        }
        await ws.send(json.dumps(context_msg))
        ctx_response = await ws.recv()
        pretty_print("Context Response", ctx_response)
        # 4. Formula execution request
        formula_msg = {
            "operation": "execute_formula",
            "data": {
                "formulaName": "mobius_collapse_v3",
                "input": [1, 2, 3, 4, 5]
            },
            "context": handshake_init["context"]
        }
        await ws.send(json.dumps(formula_msg))
        formula_response = await ws.recv()
        pretty_print("Formula Response", formula_response)
        await ws.close()
    except Exception as e:
        print('WebSocket connection: FAIL', e)

if __name__ == "__main__":
    asyncio.run(test_ws())
