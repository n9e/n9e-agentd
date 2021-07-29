package agentdctl

func inStrings(a string, b []string) bool {
	for _, v := range b {
		if v == a {
			return true
		}
	}
	return false
}

func SplitArgs(args []string, argsLenAtDash int) ([]string, []string) {
	if argsLenAtDash >= 0 && argsLenAtDash < len(args) {
		return args[:argsLenAtDash], args[argsLenAtDash:]
	}
	return args, nil
}
