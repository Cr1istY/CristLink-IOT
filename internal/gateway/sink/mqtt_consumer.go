package sink

import (
	"CristLink-IoT/internal/gateway"
	"CristLink-IoT/internal/logger"
	"encoding/json"
	"net"
	"sync"
	"time"

	MQTT "github.com/eclipse/paho.mqtt.golang"
	"go.uber.org/zap"
)

// 在设计之初，本来应该没有这个类的，但是，为了偷懒，我设计了这个类
// 因此，我可以免去自己实现MQTT协议的过程
// 但是，如果使用 其他的 MQTT Broker 来做桥接
// gnet 的性能优势就不能很好的体现出来
// 因此，这知识一个折中方案，用来在尚未自己实现 MQTT 协议解析的情况下，作为一个临时的解析MQTT到解决方案
// 然而，最终还是得在 mqtt_codec.go 中进行 MQTT 的实现

// TCPForwarder 是一个 TCP 代理，用于将 MQTT 消息变为JSON格式，然后转发给 gnet 处理
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
			logger.Logger.Debug("MQTT broker receive message: ", zap.String("topic", message.Topic()), zap.String("payload", string(message.Payload())))
			var messageData struct {
				Topic string `json:"topic"`
				Data  []byte `json:"data"`
			}
			messageData.Topic = message.Topic()
			messageData.Data = message.Payload()
			data, err := json.Marshal(messageData)
			if err != nil {
				logger.Logger.Error("json marshal error: ", zap.Error(err))
				return
			}
			forwarder.PushMessage(data)
		})
		// 异步处理订阅消息，不要阻塞 OnConnect
		go func() {
			token.Wait() // 在独立的 Goroutine 中阻塞
			if token.Wait() && token.Error() != nil {
				logger.Logger.Error("subscribe failed: ", zap.Error(token.Error()))
			}
			logger.Logger.Info("subscribe success", zap.String("topic", config.Topic))
		}()
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
	var tempDelay time.Duration // 用于重连时的指数退避

	for f.isRunning {
		conn, err := net.Dial("tcp", f.config.TCPTarget)
		if err != nil {
			logger.Logger.Error("connect to tcp target failed: ", zap.Error(err))
			time.Sleep(5 * time.Millisecond)
			continue
		}

		f.mu.Lock()
		f.conn = conn
		f.mu.Unlock()

		tempDelay = 0
		logger.Logger.Info("connect to tcp target success")

		for f.isRunning {
			if f.conn == nil {
				break
			}
			select {
			case msgBytes, ok := <-f.msgChan:
				if !ok {
					return
				}

				// 再次检查连接状态（双重检查）
				f.mu.Lock()
				currentConn := f.conn

				if currentConn == nil {
					// 连接被清空了，跳出内层循环去重连
					f.mu.Unlock()
					break
				}
				_, err = f.conn.Write(msgBytes)
				f.mu.Unlock()

				if err != nil {
					logger.Logger.Error("write to tcp target failed: ", zap.Error(err))
					err := currentConn.Close()
					if err != nil {
						logger.Logger.Error("close tcp connection failed: ", zap.Error(err))
					}

					f.mu.Lock()
					// 重连后，清理原先的f.conn
					if f.conn == currentConn {
						f.conn = nil
					}
					f.mu.Unlock()

					break
				}
			}

		}

		if conn != nil {
			_ = conn.Close()
		}
		// 如果是用户主动停止，则取消无限重试
		if !f.isRunning {
			return
		}

		// 自动重连，指数退避
		if tempDelay == 0 {
			tempDelay = time.Millisecond * 10
		} else {
			tempDelay *= 2
		}
		if maxTime := time.Second; tempDelay > maxTime {
			tempDelay = maxTime
		}

		logger.Logger.Warn("reconnecting to tcp target...", zap.Duration("sleep", tempDelay), zap.String("target", f.config.TCPTarget))
		time.Sleep(tempDelay)

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
	// TODO: 对data进行预处理

	select {
	case f.msgChan <- data:
	default:
		logger.Logger.Warn("msgChan is full, drop message")
	}
}
