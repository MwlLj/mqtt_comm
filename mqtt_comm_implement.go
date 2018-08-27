package mqtt_comm

import (
	"crypto/tls"
	"errors"
	"github.com/MwlLj/mqtt_comm/randtool"
	// "fmt"
	MQTT "github.com/eclipse/paho.mqtt.golang"
	"strconv"
	"strings"
	"sync"
	"time"
)

type CSubscribeInfo struct {
	topic string
	qos   byte
}

var m_chanMap sync.Map
var m_handleMap sync.Map
var m_subscribeTopics sync.Map
var m_serverName string
var m_serverVersion string
var m_client MQTT.Client
var m_recvQos int
var m_this *CMqttCommImplement

type CMqttCommImplement struct {
	m_connOption *MQTT.ClientOptions
}

func (this *CMqttCommImplement) Init(serverName string, versionNo string, recvQos int) {
	m_serverName = serverName
	m_serverVersion = versionNo
	m_recvQos = recvQos
	m_this = this
}

func (this *CMqttCommImplement) SetMessageBus(host string, port int, username string, userpwd string) {
	server := strings.Join([]string{"tcp://", host, ":", strconv.Itoa(port)}, "")
	// clientId := randtool.GetOrderRandStr("clientid")
	clientId := m_serverName
	this.m_connOption = MQTT.NewClientOptions().AddBroker(server).SetClientID(clientId).SetCleanSession(true)
	if username != "" {
		this.m_connOption.SetUsername(username)
		if userpwd != "" {
			this.m_connOption.SetPassword(userpwd)
		}
	}
	tlsConfig := &tls.Config{InsecureSkipVerify: true, ClientAuth: tls.NoClientCert}
	this.m_connOption.SetTLSConfig(tlsConfig)
}

func (this *CMqttCommImplement) Connect(isReConnect bool) {
	if isReConnect {
		for {
			if m_client == nil || m_client.IsConnected() == false {
				// fmt.Println("isconnect is false")
				this.connect()
			} else {
				// fmt.Println("isconnect is true")
				time.Sleep(200 * time.Millisecond)
			}
		}
	} else {
		this.connect()
	}
}

func (this *CMqttCommImplement) connect() {
	this.subscribe()
	token := m_client.Connect()
	token.Wait()
}

func onSubscribeMessage(client MQTT.Client, message MQTT.Message) {
	topic := message.Topic()
	// fmt.Println("onSubscribeMessage: " + topic)
	serverVersion, serverName, action, id, top := SpliteFullUri(topic)
	if serverName == m_serverName && serverVersion == m_serverVersion {
		// recv response
		go func() {
			// fmt.Println("get id: " + id)
			v, r := m_chanMap.Load(id)
			if !r {
				// fmt.Println("chanMap not found")
				return
			}
			ch := v.(chan string)
			ch <- string(message.Payload())
		}()
	} else {
		go func() {
			// fmt.Printf("recv subscribe message, topic: %s, payload: %s \n", topic, message.Payload())
			// recv subscribe topic
			v, r := m_handleMap.Load(top)
			// fmt.Println("get: " + top)
			if !r {
				// fmt.Println("handler map not found")
				return
			}
			handler := v.(CHandler)
			response, err := handler.Handle(top, string(message.Payload()), m_this)
			if err != nil {
				return
			}
			// fmt.Println("send response: " + GetResponseTopic(m_serverVersion, m_serverName, action, id))
			// m_client.Publish(GetResponseTopic(m_serverVersion, m_serverName, action, id), byte(qos), false, response)
			m_client.Publish(GetResponseTopic(serverVersion, serverName, action, id), byte(m_recvQos), false, response)
		}()
	}
}

func (this *CMqttCommImplement) subscribe() {
	this.m_connOption.OnConnect = func(c MQTT.Client) {
		m_subscribeTopics.Range(func(k, v interface{}) bool {
			value := v.(CSubscribeInfo)
			token := c.Subscribe(k.(string), value.qos, onSubscribeMessage)
			token.Wait()
			return true
		})
		token := c.Subscribe(GetResponseUri(m_serverVersion, m_serverName), byte(1), onSubscribeMessage)
		token.Wait()
	}
	m_client = MQTT.NewClient(this.m_connOption)
}

func (this *CMqttCommImplement) Subscribe(action string, topic string, qos int, handler CHandler) error {
	length := len(topic)
	end := []byte(topic)[length-1]
	if string(end) != "/" {
		topic += "/"
	}
	m_handleMap.Store(topic, handler)
	// fmt.Println("push back: " + topic)
	top := GetSubscribeUri(action, topic)
	subscribeInfo := CSubscribeInfo{topic: top, qos: byte(qos)}
	m_subscribeTopics.Store(top, subscribeInfo)
	// fmt.Println("subscribe: " + top)
	return nil
}

func (this *CMqttCommImplement) UnSubscribe(action string, topic string) error {
	length := len(topic)
	end := []byte(topic)[length-1]
	if string(end) != "/" {
		topic += "/"
	}
	m_handleMap.Delete(topic)
	top := GetSubscribeUri(action, topic)
	m_subscribeTopics.Delete(top)
	return nil
}

func (this *CMqttCommImplement) Send(action string, topic string, request string, qos int, timeout int) (response string, err error) {
	id := randtool.GetOrderRandStr("sessionid")
	ch := make(chan string)
	m_chanMap.Store(id, ch)
	// fmt.Println("push back id: " + id)
	// fmt.Println("send topic: " + GetFullUri(m_serverVersion, m_serverName, action, topic, id))
	m_client.Publish(GetFullUri(m_serverVersion, m_serverName, action, topic, id), byte(qos), false, request)
	select {
	case response := <-ch:
		// fmt.Println("delete id: " + id)
		m_chanMap.Delete(id)
		return response, nil
	case <-time.After(time.Duration(timeout) * time.Second):
		// fmt.Println("timeout delete id: " + id)
		m_chanMap.Delete(id)
		return "", errors.New("timeout")
	}
	return "", nil
}

func (this *CMqttCommImplement) Get(topic string, request string, qos int, timeout int) (response string, err error) {
	return this.Send("GET", topic, request, qos, timeout)
}

func (this *CMqttCommImplement) Post(topic string, request string, qos int, timeout int) (response string, err error) {
	return this.Send("POST", topic, request, qos, timeout)
}

func (this *CMqttCommImplement) Put(topic string, request string, qos int, timeout int) (response string, err error) {
	return this.Send("PUT", topic, request, qos, timeout)
}

func (this *CMqttCommImplement) Delete(topic string, request string, qos int, timeout int) (response string, err error) {
	return this.Send("DELETE", topic, request, qos, timeout)
}

func (this *CMqttCommImplement) Updated(topic string, request string, qos int) error {
	this.Send("UPDATED", topic, request, qos, 0)
	return nil
}

func (this *CMqttCommImplement) Deleted(topic string, request string, qos int) error {
	this.Send("DELETED", topic, request, qos, 0)
	return nil
}
