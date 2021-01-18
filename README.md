# Apidor
Tool for automating the search for Insecure Direct Object Reference (IDOR) vulnerabilities in web applications and APIs.

Common payloads for uncovering IDOR vulnerabilities are created using a definition file which describes the API to be tested. The payloads are then sent to the corresponding endpoints, and any unexpected responses are highlighted for further investigation.

## Definition file
The definition file is a YAML file that describes the API to be tested, as well as giving high and low privileged user tokens and other metadata. An example of a definition file is given below.

```
---
base: https://api.bncrypted.com
auth:
  header_name: Authorization
  header_value_prefix: Bearer
  high_privileged_access_token: f8h3qhwe48fh7w3874
  low_privileged_access_token: masdckjwf7h4iuhwqy3
vars:
  userId:
    high: 10001
    low: 10002
    alias: id
api:
  methods: [GET, POST, PUT, PATCH, DELETE, OPTIONS]
  headers:
  endpoints:
    users:
      - method: DELETE
        is_delete: true
        content_type: JSON
        headers:
        request_params:
        body_params:
          id: $userId
    users/$userId:
      - method: GET
        is_delete: false
        content_type: JSON
        headers:
        request_params:
        body_params:
```

 * `base`: The base URL for the API
 * `auth.header_name`: The name of the HTTP header for the authentication token
 * `auth.header_value_prefix`: The prefix for the authentication token
 * `auth.high_privileged_access_token`: The authentication token for the user with the highest privilege (aka. the victim)
 * `auth.high_privileged_access_token`: The authentication token for the user with low privilege (aka. the attacker)
 * `vars`: Variable values for high and low privileged requests to the same endpoint, accessed later in the file using `$<var_name>` syntax. For example, the high privileged user id of 10001 and the low privileged user id of 10002 can be accessed using `$userId`
 * `api.methods`: Global HTTP methods used by all API endpoints
 * `api.headers`: Global HTTP headers used by all API endpoints
 * `api.endpoints.<endpoint_name>.method`: HTTP method for that endpoint
 * `api.endpoints.<endpoint_name>.is_delete`: Whether the endpoint defines a delete instruction - the happy path test will not be run on this endpoint to preserve the resource (default: false)
 * `api.endpoints.<endpoint_name>.content_type`: The content type of the data in the request body (application/json can be represented as `JSON`; application/x-www-form-urlencoded can be represented as `FORM-DATA`)
 * `api.endpoints.<endpoint_name>.headers`: HTTP headers for that endpoint
 * `api.endpoints.<endpoint_name>.request_params`: HTTP request parameters for that endpoint
 * `api.endpoints.<endpoint_name>.body_params`: HTTP body parameters for that endpoint

## Usage
```
$ go get github.com/bncrypted/apidor
```
```
$ apidor -d <definition_file>
```
```
$ apidor -h

  Usage of apidor:
  -cert string
    	Path to a local certificate authority file
  -d string
    	Path to the API definition YAML file (default "definitions/sample.yml")
  -debug
    	Specifies whether to use debugging mode for verbose output
  -e string
    	Specifies a single endpoint operation to test (default "all")
  -ignore-base-case
    	Specifies whether to ignore base case failure
  -o string
    	Log file name
  -proxy string
    	Gives a URI to proxy HTTP traffic through
  -rate int
    	Specifies maximum number of requests made per second (default 5)
  -tests string
    	Specifies which tests should be executed (default "all")
```

## Test codes
Specific tests and combinations of tests can be specified using the `tests` flag (accepts single string or array of strings).

| Code | Test | Description | Example |
| ---- | ---- | ----------- | ------- |
| `hp` | High privileged | Tests using the high privileged variables with the high privileged access token (happy path test) | `GET /users/10001` |
| `lp` | Low privileged | Tests all variable permutations using the low privileged access token | `GET /users/10001` |
| `np` | No privileged | Tests all variable permutations without an access token | `GET /users/10001`<br>`GET /users/10002` |
| `rpp` | Request parameter pollution | Tests using parameter pollution on the request parameters | `GET /users?id=10001&id=10002` |
| `bpp` | Body parameter pollution | Tests using parameter pollution on the body parameters | `GET /users`<br><br>`id=10001&id=10002` |
| `mr` | Method replacement | Tests endpoints using unexpected HTTP methods | `PATCH /users/10001` |
| `rpw` | Request parameter wrapping | Tests using parameter wrapping on the request parameters | `GET /users?id=[10001]`<br>`GET /users?id={id:10001}` |
| `bpw` | Body parameter wrapping | Tests using parameter wrapping on the body parameters | `POST /users`<br><br>`{id: [10001]}` |
| `rps` | Request parameter substitution | Tests using body parameters substituted as request parameters | `POST /users?id=10001`<br><br>`{id: 10002}` |
| `rpspp` | Request parameter substitution with parameter pollution | Tests using body parameters substituted as request parameters alongside parameter pollution on the substituted request parameters | `POST /users?id=10001&id=10002`<br><br>`{id: 10002}` |
| `json` | JSON extension | Tests using a .json extension on the URL | `GET /users/10001.json` |
| `all` | All | Runs all tests | |
