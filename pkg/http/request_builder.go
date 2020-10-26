package http

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"
)

func buildURI(baseURI string, endpoint string, requestParams map[string]string) string {
	uri := ""

	if baseURI[len(baseURI)-1] == '/' && endpoint[0] == '/' {
		// removes potential double slash between baseURI and endpoint
		uri += baseURI + endpoint[1:]
	} else if baseURI[len(baseURI)-1] != '/' && endpoint[0] != '/' {
		// adds potential missing slash between baseURI and endpoint
		uri += baseURI + "/" + endpoint
	} else {
		uri += baseURI + endpoint
	}

	if len(requestParams) > 0 {
		orderedKeys := []string{}
		for key := range requestParams {
			orderedKeys = append(orderedKeys, key)
		}
		sort.Strings(orderedKeys)

		uri += "?"
		for _, key := range orderedKeys {
			keyWithoutIndex := key
			// checks for parameter pollution notation (eg. key:2)
			if key[len(key)-2] == ':' && (key[len(key)-1]-'0' >= 0 && key[len(key)-1]-'0' <= 9) {
				keyWithoutIndex = key[:len(key)-2]
			}
			uri += keyWithoutIndex + "=" + requestParams[key] + "&"
		}
		uri = uri[:len(uri)-1]
	}

	return uri
}

func buildBody(contentType string, bodyParams map[string]interface{}) ([]byte, error) {
	var body []byte
	var err error

	if contentType == "JSON" || strings.Contains(contentType, "application/json") {
		body, err = json.Marshal(bodyParams)
		if err != nil {
			return nil, err
		}
		// empty body
		if len(body) == 2 {
			body = []byte{}
		}
	} else if contentType == "FORM-DATA" || strings.Contains(contentType, "application/x-www-form-urlencoded") {
		body = buildFormDataBody(bodyParams)
	} else {
		if data, ok := bodyParams["data"]; ok {
			dataStr := fmt.Sprintf("%v", data)
			body = []byte(dataStr)
		} else if data, ok := bodyParams["data:1"]; ok {
			data1Str := fmt.Sprintf("%v", data)
			data2Str := fmt.Sprintf("%v", bodyParams["data:2"])
			body = []byte(data1Str + data2Str)
		} else {
			body = []byte{}
		}
	}

	indexedVarKeys := findIndexedVarKeysInBodyParams(bodyParams)
	for _, varKey := range indexedVarKeys {
		bytesToFind := []byte(varKey)
		bytesToReplace := bytesToFind[:len(bytesToFind)-2]
		body = bytes.Replace(body, bytesToFind, bytesToReplace, 1)
	}

	return body, nil
}

func buildFormDataBody(bodyParams map[string]interface{}) []byte {
	params := ""
	for paramName, paramValue := range bodyParams {
		paramStr := paramName
		if paramValue != nil {
			paramValueStr := fmt.Sprintf("%v", paramValue)
			paramStr += "=" + paramValueStr
		}
		paramStr += "&"
		params += paramStr
	}
	if len(params) > 0 {
		params = params[:len(params)-1]
	}
	return []byte(params)
}

func findIndexedVarKeysInBodyParams(bodyParams map[string]interface{}) []string {
	varKeys := []string{}
	for key, value := range bodyParams {
		switch value.(type) {
		case map[string]interface{}:
			varKeys = append(varKeys, findIndexedVarKeysInBodyParams(value.(map[string]interface{}))...)
		default:
			if len(key) >= 2 && (key[len(key)-2] == ':' && key[len(key)-1]-'0' >= 0 &&
				key[len(key)-1]-'0' <= 9) {
				varKeys = append(varKeys, key)
			}
		}
	}
	return varKeys
}

func buildContentType(contentType string) string {
	switch contentType {
	case "JSON":
		return "application/json"
	case "FORM-DATA":
		return "application/x-www-form-urlencoded"
	}
	return contentType
}

func buildClient() (*http.Client, error) {
	timeout := time.Duration(5 * time.Second)
	transport, err := buildClientTransport()
	if err != nil {
		return nil, err
	}

	client := &http.Client{
		Timeout:   timeout,
		Transport: transport,
	}

	return client, nil
}

func buildClientTransport() (*http.Transport, error) {
	var proxyURL *url.URL
	var tlsClientConfig *tls.Config
	var err error

	if isProxy {
		proxyURL, err = buildProxyURL(proxyURI)
		if err != nil {
			return nil, err
		}
	}

	if isLocalCert {
		rootCAs, err := buildRootCAs(localCertFile)
		if err != nil {
			return nil, err
		}
		tlsClientConfig = &tls.Config{
			RootCAs: rootCAs,
		}
	}

	transport := &http.Transport{}
	if isProxy && isLocalCert {
		transport = &http.Transport{
			Proxy:           http.ProxyURL(proxyURL),
			TLSClientConfig: tlsClientConfig,
		}
	} else if isProxy {
		transport = &http.Transport{
			Proxy: http.ProxyURL(proxyURL),
		}
	} else if isLocalCert {
		transport = &http.Transport{
			TLSClientConfig: tlsClientConfig,
		}
	}

	return transport, nil
}

func buildProxyURL(proxyURI string) (*url.URL, error) {
	proxyURL, err := url.Parse(proxyURI)
	if err != nil {
		return nil, errors.New("Could not parse proxy URI")
	}
	return proxyURL, nil
}

func buildRootCAs(certFile string) (*x509.CertPool, error) {
	rootCAs, _ := x509.SystemCertPool()
	if rootCAs == nil {
		rootCAs = x509.NewCertPool()
	}

	cert, err := ioutil.ReadFile(certFile)
	if err != nil {
		return nil, err
	}

	if ok := rootCAs.AppendCertsFromPEM(cert); !ok {
		return nil, errors.New("Could not append cert, using system certs only")
	}

	return rootCAs, nil
}
