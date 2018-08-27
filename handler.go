package mqtt_comm

type CHandler interface {
	Handle(topic string, request string, mc CMqttComm, user interface{}) (response string, err error)
}
