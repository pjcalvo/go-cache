package message

const (
	topic      = "deletions"
	configFile = "kafka.properties"
)

type Message struct {
	MessageID string
}
