package test

import (
	"fmt"
	"log"
	"os"
	"testing"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

// 定义 MQTT 连接参数
var f mqtt.MessageHandler = func(client mqtt.Client, msg mqtt.Message) {
	// 这里的逻辑会在收到消息时触发
	fmt.Printf("[收到消息]\n")
	fmt.Printf("  主题: %s\n", msg.Topic())
	fmt.Printf("  载荷: %s\n", string(msg.Payload()))
	fmt.Printf("  QoS: %d\n", msg.Qos())
	fmt.Printf("  保留: %v\n", msg.Retained())
}

func TestMqtt(t *testing.T) {
	// --- 1. 配置日志（可选，用于调试连接问题）---
	// 根据参考文档，你可以开启详细日志来排查问题
	mqtt.DEBUG = log.New(os.Stdout, "", 0)
	mqtt.ERROR = log.New(os.Stdout, "", 0)

	// --- 2. 创建客户端选项 ---
	opts := mqtt.NewClientOptions()
	opts.AddBroker("tcp://localhost:1883") // 替换为你的 MQTT 服务器地址
	// 如果需要用户名密码：
	// opts.SetUsername("your_username")
	// opts.SetPassword("your_password")

	// 设置客户端 ID（必须唯一，不能与服务器上其他连接的客户端重复）
	opts.SetClientID("go-subscriber-client-01")

	// 设置遗嘱消息（可选）
	// opts.SetWill("status/client", "disconnected", 1, true)

	// 设置自动重连
	opts.SetAutoReconnect(true)

	// 设置连接超时
	opts.SetConnectTimeout(10 * time.Second)

	// --- 3. 设置消息回调处理器 ---
	// 这是核心部分，参考文档中强调必须设置 Handler 来处理消息
	// 否则消息会堆积导致连接断开
	opts.SetDefaultPublishHandler(f)

	// --- 4. 创建并启动客户端 ---
	client := mqtt.NewClient(opts)

	// 连接 MQTT 服务器
	token := client.Connect()
	token.Wait() // 等待连接完成

	if token.Error() != nil {
		log.Fatal("MQTT 连接失败: ", token.Error())
	}
	fmt.Println("✅ 已成功连接到 MQTT 服务器")

	// --- 5. 订阅主题 ---
	// 这里订阅 "test/topic"，QoS 级别为 0
	topic := "/sensor/#"
	qos := byte(0)

	// 注意：参考文档中提到，如果 CleanSession 为 false，订阅前可能有旧消息到达
	// 所以必须在 Connect 之后、Subscribe 之前确保 Handler 已经设置
	subToken := client.Subscribe(topic, qos, nil)
	subToken.Wait()

	if subToken.Error() != nil {
		log.Fatal("订阅主题失败: ", subToken.Error())
	}
	fmt.Printf("✅ 已订阅主题: %s\n", topic)
	fmt.Println("等待消息输入... (按 Ctrl+C 退出)")

	// --- 6. 保持程序运行 ---
	// 这里使用一个无限循环来保持主程序不退出
	// 实际生产环境中，这里通常会是你的业务逻辑
	for {
		time.Sleep(1 * time.Second)
	}
}
