package mqtt_comm

var (
	POST      string = "POST"
	PUT       string = "PUT"
	GET       string = "GET"
	DELETE    string = "DELETE"
	UPDATED   string = "UPDATED"
	DELETED   string = "DELETED"
	actionALL string = "+"
	topicAll  string = "#"
)

type CMqttComm interface {
	Connect(isReConnect bool)
	SetMessageBus(host string, port int, username string, userpwd string)
	SubscribeAll(extraField *string, qos int, handler CHandler, user interface{}) error
	Subscribe(action string, topic string, qos int, handler CHandler, user interface{}) error
	UnSubscribe(action string, topic string) error
	IsConnect() bool
	Send(action string, topic string, request string, qos int, timeout int) (response string, err error)
	Get(topic string, request string, qos int, timeout int) (response string, err error)
	Post(topic string, request string, qos int, timeout int) (response string, err error)
	Put(topic string, request string, qos int, timeout int) (response string, err error)
	Delete(topic string, request string, qos int, timeout int) (response string, err error)
	Updated(topic string, request string, qos int) error
	Deleted(topic string, request string, qos int) error
}

func NewMqttComm(serverName string, versionNo string, recvQos int) CMqttComm {
	mqttComm := &CMqttCommImplement{}
	mqttComm.Init(serverName, versionNo, recvQos)
	return mqttComm
}
