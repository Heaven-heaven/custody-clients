package main

import (
	"crypto/rand"
	"fmt"
)
import "time"
import "io"
import "strings"
import "crypto/hmac"
import "sort"
import "net/url"
import "io/ioutil"
import "encoding/hex"
import "crypto/sha256"
import "github.com/btcsuite/btcd/btcec"
import "net/http"

const API_KEY = "x"
const API_SECRET = "x"
const HOST = "https://api.sandbox.cobo.com"

func SortParams(params map[string]string) string {
	keys := make([]string, len(params))
	i := 0
	for k, _ := range params {
		keys[i] = k
		i++
	}
	sort.Strings(keys)
	sorted := make([]string, len(params))
	i = 0
	for _, k := range keys {
		sorted[i] = k + "=" + url.QueryEscape(params[k])
		i++
	}
	return strings.Join(sorted, "&")
}
func Hash256(s string) string {
	hash_result := sha256.Sum256([]byte(s))
	hash_string := string(hash_result[:])
	return hash_string
}
func Hash256x2(s string) string {
	return Hash256(Hash256(s))
}
func SignEcc(message string) string {
	api_secret, _ := hex.DecodeString(API_SECRET)
	privKey, _ := btcec.PrivKeyFromBytes(btcec.S256(), api_secret)
	sig, _ := privKey.Sign([]byte(Hash256x2(message)))
	return fmt.Sprintf("%x", sig.Serialize())
}

func VerifyEcc(message string, signature string) bool {
	api_key, _ := hex.DecodeString(API_KEY)
	pubKey, _ := btcec.ParsePubKey(api_key, btcec.S256())

	sigBytes, _ := hex.DecodeString(signature)
	sigObj, _ := btcec.ParseSignature(sigBytes, btcec.S256())

	verified := sigObj.Verify([]byte(Hash256x2(message)), pubKey)
	return verified
}

func Request(method string, path string, params map[string]string) string {
	client := &http.Client{}
	nonce := fmt.Sprintf("%d", time.Now().Unix()*1000)
	sorted := SortParams(params)
	var req *http.Request
	if method == "POST" {
		req, _ = http.NewRequest(method, HOST+path, strings.NewReader(sorted))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	} else {
		req, _ = http.NewRequest(method, HOST+path+"?"+sorted, strings.NewReader(""))
	}
	content := strings.Join([]string{method, path, nonce, sorted}, "|")

	req.Header.Set("Biz-Api-Key", API_KEY)
	req.Header.Set("Biz-Api-Nonce", nonce)
	req.Header.Set("Biz-Api-Signature", SignEcc(content))

	resp, _ := client.Do(req)

	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)
	return string(body)
}

func main() {
	res := Request("GET", "/v1/custody/org_info/", map[string]string{})
	fmt.Printf("res %v", res)
	res = Request("GET", "/v1/custody/transaction_history/", map[string]string{
		"coin": "ETH",
		"side": "deposit",
	})
	fmt.Printf("res %v", res)
}
