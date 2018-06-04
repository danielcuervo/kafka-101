package kafka

import (
    "context"
    "encoding/json"
    "log"
    "net"
    "time"

    "github.com/Shopify/sarama"
    cluster "github.com/bsm/sarama-cluster"
)

type kafkaClient struct {
    address     string
    consumer    sarama.PartitionConsumer
    receivedMsg chan Message
}

// Messages contain domain information and should be immutable objects
type Message interface {
    Topic() string
    Payload() map[string]interface{}
}

// Creates a driver that consumes kafka messages
func NewClient(address string) (*kafkaClient, error) {
    ensureServicesAreAlive()
    return &kafkaClient{address: address, receivedMsg: make(chan Message)}, nil
}

func ensureServicesAreAlive() {
    for {
        conn, err := net.DialTimeout("tcp", "kafka:9092", time.Second)
        if conn != nil {
            return
        }

        log.Println(err.Error())
        time.Sleep(time.Second * 10)
    }
}

func (kc *kafkaClient) Receive() <-chan Message {
    return kc.receivedMsg
}

func (kc *kafkaClient) Consume(topic string, partition int32, groupID string, ctx context.Context) error {
    consumer, err := cluster.NewConsumer(
        []string{kc.address},
        groupID,
        []string{topic},
        cluster.NewConfig(),
    )
    if err != nil {
        log.Println(err)
        return err
    }

    for {
        select {
        case err := <-consumer.Errors():
            log.Println(err.Error())
            return nil
        case <-ctx.Done():
            return nil
        case msg := <-consumer.Messages():
            consumer.MarkOffset(msg, groupID)
            payload := &map[string]interface{}{}
            json.Unmarshal(msg.Value, payload)
            kc.receivedMsg <- &message{
                topic:   msg.Topic,
                payload: *payload,
            }
        }
    }

    return nil
}

func (kc *kafkaClient) Dispatch(msg Message) error {
    producer, err := sarama.NewAsyncProducer([]string{kc.address}, sarama.NewConfig())
    if err != nil {
        return err
    }
    producer.Input() <- &sarama.ProducerMessage{
        Topic: msg.Topic(),
        Value: &payloadEncoder{msg.Payload()},
    }

    return nil
}

type payloadEncoder struct {
    Payload map[string]interface{}
}

func (pe *payloadEncoder) Length() int {
    encoded, err := json.Marshal(pe.Payload)
    if err != nil {
        return 0
    }

    return len(encoded)
}

func (pe *payloadEncoder) Encode() ([]byte, error) {
    encoded, err := json.Marshal(pe.Payload)
    return encoded, err
}

type message struct {
    topic   string
    payload map[string]interface{}
}

func (m *message) Topic() string {
    return m.topic
}

func (m *message) Payload() map[string]interface{} {
    return m.payload
}
