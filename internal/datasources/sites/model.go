package sites

type Site struct {
	Name        string `tfsdk:"name"`
	Code        string `tfsdk:"code"`
	Description string `tfsdk:"description"`
	Location    string `tfsdk:"location"`
}

type State struct {
	ID    string `tfsdk:"id"`
	Sites []Site `tfsdk:"sites"`
}
