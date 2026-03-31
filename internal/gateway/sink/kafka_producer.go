package sink

import (
	"context"
	"log"

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
		Addr:     kafka.TCP(addr),
		Topic:    topic,
		Balancer: &kafka.LeastBytes{}, // 简单的负载均衡策略
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
	// MVP 阶段使用 WriteMessages，生产环境建议异步批量写入以优化性能
	if err := p.writer.WriteMessages(ctx, msg); err != nil {
		log.Printf("Kafka write error: %v", err)
		return err
	}
	log.Println("kafka write success")
	return nil
}

// Close 关闭连接
func (p *Producer) Close() error {
	return p.writer.Close()
}
