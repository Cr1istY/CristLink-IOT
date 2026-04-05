import random
import time
import json
import paho.mqtt.client as mqtt

# brew install mosquitto
# 使用 mosquitto 作为本地 MQTT Broker
# 开启 匿名访问模式
# Topic 通常包含路由信息（如 /sys/{pk}/{dk}/up）

# 配置区域
BROKER_HOST = "localhost"
BROKER_PORT = 9001

DEVICE_ID = "device_1"
PRODUCT_KEY = "pk_test_123"

TOPIC = f"/sys/{PRODUCT_KEY}/{DEVICE_ID}/up"

def on_connect(client, userdata, flags, rc, properties):
    if rc == 0:
        print("Connected successfully.")
    else:
        print(f"Failed to connect, return code {rc}\n")

def on_disconnect(client, userdata, rc):
    print("Disconnected.")

def send_message():
    client = mqtt.Client(callback_api_version=mqtt.CallbackAPIVersion.VERSION2,
                         client_id=f"python_sim_{DEVICE_ID}")
    client.on_connect = on_connect
    client.on_disconnect = on_disconnect

    print(f"Connecting to {BROKER_HOST}:{BROKER_PORT}...")
    client.connect(BROKER_HOST, BROKER_PORT, 60)
    client.loop_start()

    try:
        print(f"Sending message to {TOPIC}...")
        while True:
            temp = round(random.uniform(20.0, 30.0), 2)
            humi = round(random.uniform(40.0, 80.0), 2)

            payload = {
                "device_id": DEVICE_ID,
                "product_key": PRODUCT_KEY,
                "ts": int(time.time() * 1000),
                "data": {
                    "temperature": temp,
                    "humidity": humi,
                    "status": "online"
                }
            }

            result = client.publish(TOPIC, json.dumps(payload), qos=1)
            status = "success" if result.rc == 0 else "failed"
            print(f"Message sent to {TOPIC} with status: {status} at ts " + str(payload["ts"]))

            time.sleep(2)

    except KeyboardInterrupt:
        print("Disconnecting from broker...")
    finally:
        client.loop_stop()
        client.disconnect()

if __name__ == "__main__":
    send_message()