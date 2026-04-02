package sink

import (
	"context"
	"log"
	"time"

	"github.com/segmentio/kafka-go"
)

// Producer 封装 Kafka 生产者
type Producer struct {
	writer *kafka.Writer
	topic  string
}

// NewProducer 初始化生产者
func NewProducer(addr, topic string) *Producer {
	w := &kafka.Writer{
		Addr:                   kafka.TCP(addr),
		Topic:                  topic,
		Balancer:               &kafka.LeastBytes{}, // 简单的负载均衡策略
		AllowAutoTopicCreation: true,                // 允许自动创建主题
	}
	return &Producer{writer: w, topic: topic}
}

// 实现数据的异步投递。为了不影响网络接收协程的性能，我们在这里做一个简单的封装。

// Send 发送数据到 Kafka
func (p *Producer) Send(ctx context.Context, key, value []byte) error {
	msg := kafka.Message{
		Key:   key,
		Value: value,
	}
	// 超时控制
	ctxWithTimeout, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := p.writer.WriteMessages(ctxWithTimeout, msg); err != nil {
		// 生产环境建议使用 zap 等高性能日志库，并增加错误分类处理
		// 例如：如果是 "queue full"，可以稍后重试；如果是 "topic not exist"，则报警
		log.Printf("Kafka write error: %v", err)
		return err
	}
	return nil
}

// Close 关闭连接
func (p *Producer) Close() error {
	return p.writer.Close()
}
