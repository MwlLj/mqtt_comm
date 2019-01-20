package mqtt_comm

import (
	"errors"
	"fmt"
	"testing"
	"time"
)

func TestMqttComm(t *testing.T) {
	mqttComm := NewMqttComm("test", "1.0", 0)
	mqttComm.SetMessageBus("127.0.0.1", 51883, "", "")
	mqttComm.Connect(false)
	go func() {
		recv, _ := mqttComm.Send("GET", "test", "hello", 1, 60)
		fmt.Println("recv: " + recv)
	}()
	for {
		time.Sleep(time.Second * 1)
	}
}

type MyHandler struct {
}

func (*MyHandler) Handle(topic *string, action *string, request *string, qos int, mc CMqttComm, user interface{}) (response *string, err error) {
	mc.Get("1.0/test", "hello", 0, 10)
	return nil, errors.New("")
}

func TestSubscribe(t *testing.T) {
	mqttComm := NewMqttComm("test", "1.0", 0)
	mqttComm.SetMessageBus("127.0.0.1", 51883, "", "")
	mqttComm.Subscribe("GET", "test", 0, &MyHandler{}, nil)
	mqttComm.Connect(true)
}
