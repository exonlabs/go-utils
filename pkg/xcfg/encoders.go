package xcfg

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
)

type Encoder interface {
	Encode(any) ([]byte, error)
	Decode([]byte) (any, error)
}

type defaultEncoder struct{}

// encoding with byte mangling
func (enc *defaultEncoder) Encode(data any) ([]byte, error) {
	b, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}
	b = _mangle([]byte(base64.StdEncoding.EncodeToString(b)))
	return b, nil
}

// encoding with byte mangling
func (enc *defaultEncoder) Decode(data []byte) (any, error) {
	if data == nil {
		return nil, nil
	}
	b, err := base64.StdEncoding.DecodeString(string(_mangle(data)))
	if err != nil {
		return nil, err
	}
	var result any
	err = json.Unmarshal(b, &result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// mangle bytes data to avoid easy detectable patterns.
// inverting every 2 bytes order
func _mangle(b []byte) []byte {
	res := bytes.Clone(b)
	// we skip indexs:0,n-1,n-2 for extra mangling ;)
	for i := 1; i < len(b)-3; i += 2 {
		res[i], res[i+1] = b[i+1], b[i]
	}
	return res
}
