package options

const (
	_ uint16 = iota << 3

	PRI_M_CORE      // no  dep
	PRI_M_APISERVER // dep  core
)
