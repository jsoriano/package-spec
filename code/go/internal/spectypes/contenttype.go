// Copyright Elasticsearch B.V. and/or licensed to Elasticsearch B.V. under one
// or more contributor license agreements. Licensed under the Elastic License;
// you may not use this file except in compliance with the Elastic License.

package spectypes

import (
	"encoding/json"
	"fmt"
	"mime"

	"gopkg.in/yaml.v3"
)

type ContentType struct {
	MediaType string
	Params    map[string]string
}

func (t ContentType) MarshalJSON() ([]byte, error) {
	return []byte(`"` + t.String() + `"`), nil
}

func (t ContentType) MarshalYAML() (interface{}, error) {
	return t.String(), nil
}

func (t *ContentType) UnmarshalJSON(d []byte) error {
	var raw string
	err := json.Unmarshal(d, &raw)
	if err != nil {
		return err
	}

	return t.unmarshalString(raw)
}

func (t *ContentType) UnmarshalYAML(value *yaml.Node) error {
	// For some reason go-yaml doesn't like the UnmarshalJSON function above.
	return t.unmarshalString(value.Value)
}

func (t *ContentType) unmarshalString(text string) error {
	mediatype, params, err := mime.ParseMediaType(text)
	if err != nil {
		return err
	}
	if mime.FormatMediaType(mediatype, params) == "" {
		// Bug in mime library? Happens when parsing something like "0;*0=0"
		return fmt.Errorf("invalid token in media type")
	}

	t.MediaType = mediatype
	t.Params = params
	return nil
}

func (t ContentType) String() string {
	return mime.FormatMediaType(t.MediaType, t.Params)
}
