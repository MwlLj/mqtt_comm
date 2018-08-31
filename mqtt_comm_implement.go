package mqtt_comm

import (
	"crypto/tls"
	"errors"
	// "fmt"
	"github.com/MwlLj/mqtt_comm/randtool"
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

type CHandlerInfo struct {
	handler CHandler
	user    interface{}
}

type CMqttCommImplement struct {
	m_chanMap         sync.Map
	m_handleMap       sync.Map
	m_subscribeTopics sync.Map
	m_serverName      string
	m_serverVersion   string
	m_client          MQTT.Client
	m_recvQos         int
	m_connOption      *MQTT.ClientOptions
}

var globThisMap sync.Map

func (this *CMqttCommImplement) Init(serverName string, versionNo string, recvQos int) {
	this.m_serverName = serverName
	this.m_serverVersion = versionNo
	this.m_recvQos = recvQos
}

func (this *CMqttCommImplement) SetMessageBus(host string, port int, username string, userpwd string) {
	server := strings.Join([]string{"tcp://", host, ":", strconv.Itoa(port)}, "")
	clientId := randtool.GetOrderRandStr("clientid")
	// clientId := m_serverName
	globThisMap.Store(clientId, this)
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
			if this.m_client == nil || this.m_client.IsConnected() == false {
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
	token := this.m_client.Connect()
	token.Wait()
}

func onSubscribeMessage(client MQTT.Client, message MQTT.Message) {
	optionReader := client.OptionsReader()
	clientId := optionReader.ClientID()
	thisValue, thisErr := globThisMap.Load(clientId)
	if !thisErr {
		// fmt.Println("this not found")
		return
	}
	this := thisValue.(*CMqttCommImplement)
	topic := message.Topic()
	// fmt.Println("onSubscribeMessage: "+topic, ", content: ", message.Payload())
	serverVersion, serverName, action, id, top, length := SpliteFullUri(topic)
	// if serverName == this.m_serverName && serverVersion == this.m_serverVersion {
	if length == 5 {
		// recv response
		go func() {
			// fmt.Println("get id: " + id)
			v, r := this.m_chanMap.Load(id)
			if !r {
				// fmt.Println("chanMap not found")
				return
			}
			ch := v.(chan string)
			ch <- string(message.Payload())
		}()
	} else if length > 5 {
		go func() {
			// fmt.Printf("recv subscribe message, topic: %s, payload: %s \n", topic, message.Payload())
			// recv subscribe topic
			v, r := this.m_handleMap.Load(top)
			// fmt.Println("get: " + top)
			if !r {
				// fmt.Println("handler map not found")
				return
			}
			handlerInfo := v.(CHandlerInfo)
			response, _ := handlerInfo.handler.Handle(top, string(message.Payload()), this, handlerInfo.user)
			// fmt.Println("send response topic: "+GetResponseTopic(serverVersion, serverName, action, id), ", response: ", response)
			this.m_client.Publish(GetResponseTopic(serverVersion, serverName, action, id), byte(this.m_recvQos), false, response)
		}()
	}
}

func (this *CMqttCommImplement) subscribe() {
	this.m_connOption.OnConnect = func(c MQTT.Client) {
		this.m_subscribeTopics.Range(func(k, v interface{}) bool {
			value := v.(CSubscribeInfo)
			token := c.Subscribe(k.(string), value.qos, onSubscribeMessage)
			token.Wait()
			return true
		})
		token := c.Subscribe(GetResponseUri(this.m_serverVersion, this.m_serverName), byte(1), onSubscribeMessage)
		// fmt.Println("subscribe: " + GetResponseUri(this.m_serverVersion, this.m_serverName))
		token.Wait()
	}
	this.m_client = MQTT.NewClient(this.m_connOption)
}

func (this *CMqttCommImplement) Subscribe(action string, topic string, qos int, handler CHandler, user interface{}) error {
	length := len(topic)
	end := []byte(topic)[length-1]
	if string(end) != "/" {
		topic += "/"
	}
	handlerInfo := CHandlerInfo{handler: handler, user: user}
	this.m_handleMap.Store(topic, handlerInfo)
	// fmt.Println("push back: " + topic)
	top := GetSubscribeUri(action, topic)
	subscribeInfo := CSubscribeInfo{topic: top, qos: byte(qos)}
	this.m_subscribeTopics.Store(top, subscribeInfo)
	// fmt.Println("subscribe: " + top)
	return nil
}

func (this *CMqttCommImplement) UnSubscribe(action string, topic string) error {
	length := len(topic)
	end := []byte(topic)[length-1]
	if string(end) != "/" {
		topic += "/"
	}
	this.m_handleMap.Delete(topic)
	top := GetSubscribeUri(action, topic)
	this.m_subscribeTopics.Delete(top)
	return nil
}

func (this *CMqttCommImplement) Send(action string, topic string, request string, qos int, timeout int) (response string, err error) {
	id := randtool.GetOrderRandStr("sessionid")
	ch := make(chan string)
	this.m_chanMap.Store(id, ch)
	// fmt.Println("push back id: " + id)
	// fmt.Println("send topic: " + GetFullUri(this.m_serverVersion, this.m_serverName, action, topic, id))
	this.m_client.Publish(GetFullUri(this.m_serverVersion, this.m_serverName, action, topic, id), byte(qos), false, request)
	select {
	case response := <-ch:
		// fmt.Println("delete id: " + id)
		this.m_chanMap.Delete(id)
		return response, nil
	case <-time.After(time.Duration(timeout) * time.Second):
		// fmt.Println("timeout delete id: " + id)
		this.m_chanMap.Delete(id)
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
