package resources

type Plan struct {
	ID        string   `tfsdk:"id"`
	Level     string   `tfsdk:"level"`
	Includes  []string `tfsdk:"includes"`
	Excludes  []string `tfsdk:"excludes"`
	Resources []Item   `tfsdk:"resources"`
}

type Item struct {
	Model string `tfsdk:"model"`
}
