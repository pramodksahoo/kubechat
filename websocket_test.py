#!/usr/bin/env python3

import websocket
import json
import threading
import time

# WebSocket testing for KubeChat API
WS_URL = "ws://localhost:30080/api/v1/ws"
TOKEN = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoiMjNjNWYzN2MtNzRmYy00NTJkLWFhZTktZTcwYmY5MTE1NjU3IiwidXNlcm5hbWUiOiJ3c3Rlc3QiLCJyb2xlIjoidXNlciIsInNlc3Npb25faWQiOiJhMjNlZWViMS1iM2ZiLTQzNjQtOWFlZC1mMzZjMWUwZGJiOGIiLCJpc3MiOiJrdWJlY2hhdC1hdXRoIiwic3ViIjoiMjNjNWYzN2MtNzRmYy00NTJkLWFhZTktZTcwYmY5MTE1NjU3IiwiZXhwIjoxNzU3Njc3NDkwLCJuYmYiOjE3NTc1OTEwOTAsImlhdCI6MTc1NzU5MTA5MH0.lAFgwwVD3lv5-Ke7spUTSKWNk46mBguiOndzr1XG_-g"

def on_message(ws, message):
    print(f"📨 Received: {message}")
    try:
        data = json.loads(message)
        print(f"📋 Message type: {data.get('type', 'unknown')}")
    except:
        print("⚠️  Not JSON format")

def on_error(ws, error):
    print(f"❌ Error: {error}")

def on_close(ws, close_status_code, close_msg):
    print(f"🔚 Connection closed: {close_status_code} - {close_msg}")

def on_open(ws):
    print("✅ WebSocket connected!")
    
    # Test authentication
    auth_msg = {
        "type": "auth",
        "payload": {
            "token": TOKEN
        }
    }
    
    print("🔐 Sending authentication...")
    ws.send(json.dumps(auth_msg))
    
    # Schedule close after testing
    def close_later():
        time.sleep(5)
        print("🔚 Closing connection...")
        ws.close()
    
    threading.Thread(target=close_later).start()

if __name__ == "__main__":
    print("🔗 Testing WebSocket connection to KubeChat API")
    print(f"📡 URL: {WS_URL}")
    
    try:
        # Enable debug for more info
        # websocket.enableTrace(True)
        ws = websocket.WebSocketApp(WS_URL,
                                  on_open=on_open,
                                  on_message=on_message,
                                  on_error=on_error,
                                  on_close=on_close)
        
        ws.run_forever()
        print("✅ WebSocket test completed")
        
    except Exception as e:
        print(f"❌ Test failed: {e}")