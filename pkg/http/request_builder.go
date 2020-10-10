package http

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"net/url"
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
		uri += "?"
		for key, param := range requestParams {
			// checks for parameter pollution notation (eg. key:2)
			if key[len(key)-2] == ':' && (key[len(key)-1]-'0' >= 0 && key[len(key)-1]-'0' <= 9) {
				key = key[:len(key)-2]
			}
			uri += key + "=" + param + "&"
		}
		uri = uri[:len(uri)-1]
	}

	return uri
}

func buildBody(encodedContentType string, bodyParams map[string]interface{}) ([]byte, error) {
	var body []byte
	var err error
	switch encodedContentType {
	case "JSON":
		body, err = json.Marshal(bodyParams)
		if err != nil {
			return nil, err
		}
	default:
		return nil, errors.New("Unknown content type \"" + encodedContentType + "\"")
	}
	return body, nil
}

func buildContentType(encodedContentType string) (string, error) {
	contentType := ""
	switch encodedContentType {
	case "JSON":
		contentType += "application/json"
	case "FORM-DATA":
		contentType += "application/x-www-form-urlencoded"
	default:
		return "", errors.New("Unknown content type \"" + encodedContentType + "\"")
	}
	return contentType, nil
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
