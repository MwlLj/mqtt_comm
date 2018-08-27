package mqtt_comm

type CHandler interface {
	Handle(topic string, request string, mc CMqttComm) (response string, err error)
}
