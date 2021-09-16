package mocker

type Config struct {
	Port        int    `flag:"port" default:"8000" env:"N9E_MOCKER_PORT" description:"listen port"`
	CollectRule bool   `flag:"collect-rule" description:"enable send statsd sample data"`
	SendStatsd  bool   `flag:"send-statsd" description:"enable collect rule provider"`
	Confd       string `flag:"confd" default:"./etc/mocker.d" description:"config dir"`
}
