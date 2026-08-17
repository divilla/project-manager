package runner

import "errors"

var CommandError = errors.New("command error")
var CurlError = errors.New("curl error")
var JQSelectorError = errors.New("jq selector error")

func Curl(method string, url string, headers map[string]string, timeout int, retries int, query string, body string) (string, int, error) {
	return "", 0, nil
}
func JQFilter(selector, input string) (string, int, error) {
	return "", 0, nil
}
