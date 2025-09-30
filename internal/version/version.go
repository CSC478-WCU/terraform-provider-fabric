package version

var v = "dev"

func Set(s string) { v = s }
func Get() string  { return v }
