// Extend https://github.com/ghodss/yaml to be able to handle multi-document
// streams.  This uses go-yaml to parse the stream, then converts back to YAML
// and feeds the result to ghodss/yaml, which unmarshals it, converts it to
// JSON, and then unmarshals that into a data structure

package yaml

import (
	"io"

	ghodssYaml "github.com/ghodss/yaml"
	goYaml "gopkg.in/yaml.v2"
)

type Decoder struct {
	decoder *goYaml.Decoder
}

func NewDecoder(r io.Reader) *Decoder {
	return &Decoder{
		decoder: goYaml.NewDecoder(r),
	}
}

func (dec *Decoder) Decode(result interface{}) (err error) {
	var singleValue interface{}

	// decode a single document from the stream
	if err := dec.decoder.Decode(&singleValue); err != nil {
		return err
	}

	// re-encode that as a single YAML string
	singleString, err := goYaml.Marshal(singleValue)
	if err != nil {
		return err
	}

	// decode that using ghodss/yaml
	err = ghodssYaml.Unmarshal(singleString, &result)
	if err != nil {
		return err
	}

	return nil
}
