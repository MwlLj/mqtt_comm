package mqtt_comm

type CHandler interface {
	Handle(topic *string, action *string, request *string, qos int, mc CMqttComm, user interface{}) (response *string, err error)
}
