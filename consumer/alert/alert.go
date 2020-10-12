package alert

//Alert interface defined ...
type Alert interface {
	PushNotify(msg string) error
}
