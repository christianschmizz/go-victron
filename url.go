package vrm

import (
	"bytes"
	"reflect"
	"text/template"

	"github.com/google/go-querystring/query"
)

type URLParams map[string]string

// MapTemplate formats the given string by applying the replacement map
func MapTemplate(templateText string, replacements map[string]string) (string, error) {
	tpl, err := template.New("").Parse(templateText)
	if err != nil {
		return "", err
	}
	var buf bytes.Buffer
	err = tpl.Execute(&buf, replacements)
	if err != nil {
		return "", err
	}
	return buf.String(), nil
}

func formatURL(urlTemplate string, params URLParams, queryParams interface{}) (string, error) {
	params["baseURL"] = baseURL

	url, err := MapTemplate(urlTemplate, params)
	if err != nil {
		return "", err
	}
	if queryParams != nil && !reflect.DeepEqual(queryParams, reflect.Zero(reflect.TypeOf(queryParams)).Interface()) {
		queryString, err := query.Values(queryParams)
		if err != nil {
			return "", err
		}
		url += "?" + queryString.Encode()
	}
	return url, nil
}