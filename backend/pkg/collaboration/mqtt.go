package collaboration

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/HardMax71/syncwrite/backend/pkg/utils"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"go.uber.org/zap"
)

type MQTTClient struct {
	client   mqtt.Client
	logger   *zap.Logger
	handlers map[string]func([]byte)
	mutex    sync.RWMutex
}

func NewMQTTClient(brokerURL string) (*MQTTClient, error) {
	logger := utils.Logger()

	opts := mqtt.NewClientOptions().
		AddBroker(brokerURL).
		SetClientID(fmt.Sprintf("syncwrite-server-%d", time.Now().UnixNano())).
		SetCleanSession(true).
		SetAutoReconnect(true).
		SetOnConnectHandler(func(client mqtt.Client) {
			logger.Info("Connected to MQTT broker")
		}).
		SetConnectionLostHandler(func(client mqtt.Client, err error) {
			logger.Error("Lost connection to MQTT broker", zap.Error(err))
		})

	client := mqtt.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		return nil, fmt.Errorf("error connecting to MQTT broker: %w", token.Error())
	}

	return &MQTTClient{
		client:   client,
		logger:   logger,
		handlers: make(map[string]func([]byte)),
	}, nil
}

func (m *MQTTClient) Subscribe(topic string, handler func([]byte)) error {
	m.mutex.Lock()
	m.handlers[topic] = handler
	m.mutex.Unlock()

	token := m.client.Subscribe(topic, 1, func(client mqtt.Client, msg mqtt.Message) {
		m.mutex.RLock()
		if handler, exists := m.handlers[msg.Topic()]; exists {
			handler(msg.Payload())
		}
		m.mutex.RUnlock()
	})

	token.Wait()
	return token.Error()
}

func (m *MQTTClient) Unsubscribe(topic string) error {
	m.mutex.Lock()
	delete(m.handlers, topic)
	m.mutex.Unlock()

	token := m.client.Unsubscribe(topic)
	token.Wait()
	return token.Error()
}

func (m *MQTTClient) Publish(topic string, payload interface{}) error {
	data, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("error marshaling payload: %w", err)
	}

	token := m.client.Publish(topic, 1, false, data)
	token.Wait()
	return token.Error()
}

func (m *MQTTClient) Close() {
	m.client.Disconnect(250)
}

func GetDocumentTopic(documentID string) string {
	return fmt.Sprintf("documents/%s/changes", documentID)
}

func GetPresenceTopic(documentID string) string {
	return fmt.Sprintf("documents/%s/presence", documentID)
}
