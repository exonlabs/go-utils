package xcipher_test

import (
	"testing"

	"github.com/exonlabs/go-utils/pkg/crypto/xcipher"
	"github.com/exonlabs/go-utils/tests"
)

func TestAES128_Encryption(t *testing.T) {
	secret := "123456"
	aes, err := xcipher.NewAES128(secret)
	if err != nil {
		t.Errorf(tests.FailMsg()+" -- error: %s", err)
		return
	}

	txt_in := "### INPUT TEXT FOR ENCRYPTION ###"
	b_ciphered, err := aes.Encrypt([]byte(txt_in))
	if err != nil {
		t.Errorf(tests.FailMsg()+" -- error: %s", err)
		return
	}

	b_out, err := aes.Decrypt(b_ciphered)
	if err != nil {
		t.Errorf(tests.FailMsg()+" -- error: %s", err)
		return
	}
	txt_out := string(b_out)

	t.Logf("input: %v ---> %v\n", txt_in, txt_out)
	if txt_in == txt_out {
		t.Logf(tests.ValidMsg())
	} else {
		t.Errorf(tests.FailMsg())
	}
}

func TestAES256_Encryption(t *testing.T) {
	secret := "123456"
	aes, err := xcipher.NewAES256(secret)
	if err != nil {
		t.Errorf(tests.FailMsg()+" -- error: %s", err)
		return
	}

	txt_in := "### INPUT TEXT FOR ENCRYPTION ###"
	b_ciphered, err := aes.Encrypt([]byte(txt_in))
	if err != nil {
		t.Errorf(tests.FailMsg()+" -- error: %s", err)
		return
	}

	b_out, err := aes.Decrypt(b_ciphered)
	if err != nil {
		t.Errorf(tests.FailMsg()+" -- error: %s", err)
		return
	}
	txt_out := string(b_out)

	t.Logf("input: %v ---> %v\n", txt_in, txt_out)
	if txt_in == txt_out {
		t.Logf(tests.ValidMsg())
	} else {
		t.Errorf(tests.FailMsg())
	}
}
