package server

// action identifies an SQS API operation as found in the Action form field.
type action string

const (
	actionCreateQueue    action = "CreateQueue"
	actionDeleteQueue    action = "DeleteQueue"
	actionListQueues     action = "ListQueues"
	actionSendMessage    action = "SendMessage"
	actionReceiveMessage action = "ReceiveMessage"
	actionDeleteMessage  action = "DeleteMessage"
)
