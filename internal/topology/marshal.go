package topology

import (
	"bytes"
	"encoding/xml"
)

func Marshal(g GraphML) (string, error) {
	buf := &bytes.Buffer{}
	buf.WriteString(xml.Header)
	enc := xml.NewEncoder(buf)
	enc.Indent("", "  ")
	if err := enc.Encode(g); err != nil {
		return "", err
	}
	return buf.String(), nil
}
