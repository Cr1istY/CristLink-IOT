package sink

import (
	"CristLink-IoT/internal/gateway"
	"CristLink-IoT/internal/logger"
	"net"
	"sync"
	"time"

	MQTT "github.com/eclipse/paho.mqtt.golang"
	"go.uber.org/zap"
)

type TCPForwarder struct {
	config    *gateway.TCPForwarderConfig
	client    MQTT.Client
	msgChan   chan []byte // 使用 Channel 进行异步解耦
	conn      net.Conn    // 保持长连接
	mu        sync.Mutex  // 写锁，防止并发写入导致数据交错
	isRunning bool
}

func NewTCPForwarder(cfg *gateway.TCPForwarderConfig) *TCPForwarder {
	config := *gateway.GetTCPForwarderConfig()
	opts := MQTT.NewClientOptions()
	opts.AddBroker(config.Addr)
	opts.SetClientID(config.ClientID)
	opts.SetUsername(config.UserName)
	opts.SetPassword(config.Password)
	opts.SetAutoReconnect(config.AutoReconnect)
	opts.SetResumeSubs(config.AutoReconnect)
	opts.SetCleanSession(config.CleanSession)

	forwarder := &TCPForwarder{
		config:    cfg, // 存入配置
		msgChan:   make(chan []byte, 1024),
		isRunning: false,
	}

	opts.OnConnect = func(client MQTT.Client) {
		logger.Info("MQTT connected")
		token := client.Subscribe(cfg.Topic, 0, func(client MQTT.Client, message MQTT.Message) {
			logger.Logger.Debug("MQTT broker receive message: ", zap.Any("topic", message.Topic()), zap.Any("payload", message.Payload()))
			data := make([]byte, len(message.Payload()))
			copy(data, message.Payload())
			forwarder.PushMessage(data)
		})

		if token.Wait() && token.Error() != nil {
			logger.Logger.Error("subscribe failed: ", zap.Error(token.Error()))
		}
		logger.Logger.Info("subscribe success", zap.String("topic", config.Topic))
	}

	opts.OnConnectionLost = func(client MQTT.Client, err error) {
		logger.Logger.Error("MQTT connection lost: ", zap.Error(err))
	}

	forwarder.client = MQTT.NewClient(opts)
	if token := forwarder.client.Connect(); token.Wait() && token.Error() != nil {
		logger.Logger.Error("create client failed", zap.Error(token.Error()))
		return nil
	}

	return forwarder
}

// Start 启动转发器 (包含自动重连逻辑)
func (f *TCPForwarder) Start() {
	f.isRunning = true
	go f.tcpWriteLoop()
}

// tcpWriteLoop TCP 写入循环
func (f *TCPForwarder) tcpWriteLoop() {
	var err error

	for f.isRunning {
		f.conn, err = net.Dial("tcp", f.config.TCPTarget)
		if err != nil {
			logger.Logger.Error("connect to tcp target failed: ", zap.Error(err))
			time.Sleep(2 * time.Second)
			continue
		}
		logger.Logger.Info("connect to tcp target success")

		for f.isRunning {
			select {
			case msgBytes, ok := <-f.msgChan:
				if !ok {
					return
				}
				f.mu.Lock()
				_, err = f.conn.Write(msgBytes)
				f.mu.Unlock()

				if err != nil {
					logger.Logger.Error("write to tcp target failed: ", zap.Error(err))
					err := f.conn.Close()
					if err != nil {
						logger.Logger.Error("close tcp connection failed: ", zap.Error(err))
					}
					break
				}
			}
		}

	}

}

func (f *TCPForwarder) Stop() {
	f.isRunning = false
	if f.conn != nil {
		err := f.conn.Close()
		if err != nil {
			logger.Logger.Error("close tcp connection failed: ", zap.Error(err))
		}
		if f.client.IsConnected() {
			f.client.Disconnect(250)
		}
		close(f.msgChan)
	}
}

// PushMessage 接收 MQTT 消息并推送到 Channel
func (f *TCPForwarder) PushMessage(data []byte) {
	select {
	case f.msgChan <- data:
	default:
		logger.Logger.Warn("msgChan is full, drop message")
	}
}
