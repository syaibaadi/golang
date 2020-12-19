package helpers

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"math"
	"math/rand"
	"strconv"
	"strings"
	"time"

	"gitlab.com/pangestu18/janji-online/chat/config"
	"github.com/mergermarket/go-pkcs7"
	"golang.org/x/crypto/bcrypt"
)

const time_diff_limit = 480

// Decrypt decrypts cipher text string into plain text string
func EncryptDecrypt(encrypted string, isEncrypt bool) (string, error) {
	appkey := config.Get("APP_KEY").String()
	appkey = strings.Replace(appkey, "base64:", "", 1)
	key := []byte(appkey)
	h := sha256.New()
	h.Write(key)
	key = h.Sum(nil)[:32]
	iv := []byte("\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00")

	q, err := mdecrypt(key, iv, encrypted)
	if err == nil {
		return string(q), err
	}
	return "", nil
}

func mdecrypt(key []byte, iv []byte, encrypted string) ([]byte, error) {
	data, err := base64.StdEncoding.DecodeString(encrypted)
	if err != nil {
		return nil, err
	}

	if len(data) == 0 || len(data)%aes.BlockSize != 0 {
		return nil, fmt.Errorf("bad blocksize(%v), aes.BlockSize = %v\n", len(data), aes.BlockSize)
	}
	c, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	cbc := cipher.NewCBCDecrypter(c, iv)
	cbc.CryptBlocks(data, data)
	out, err := pkcs7.Unpad(data, aes.BlockSize)
	if err != nil {
		return out, err
	}
	return out, nil
}

func tsDiff(ts string) bool {
	_ts, err := strconv.ParseInt(ts, 10, 64)
	if err != nil {
		return false
	}
	return math.Abs(float64(_ts-time.Now().Unix())) <= time_diff_limit
}

func doubleEncrypt(str string, cid string, sck string) string {
	arr := []byte(str)
	result := encrypt(arr, cid)
	result = encrypt(result, sck)
	return strings.Replace(strings.Replace(strings.TrimRight(base64.StdEncoding.EncodeToString(result), "="), "+", "-", -1), "/", "_", -1)
}

func encrypt(str []byte, k string) []byte {
	var result []byte
	strls := len(str)
	strlk := len(k)
	for i := 0; i < strls; i++ {
		char := str[i]
		keychar := k[(i+strlk-1)%strlk]
		char = byte((int(char) + int(keychar)) % 128)
		result = append(result, char)
	}
	return result
}

func doubleDecrypt(str string, cid string, sck string) string {
	if i := len(str) % 4; i != 0 {
		str += strings.Repeat("=", 4-i)
	}
	result, err := base64.StdEncoding.DecodeString(strings.Replace(strings.Replace(str, "-", "+", -1), "_", "/", -1))
	if err != nil {
		return ""
	}
	result = decrypt(result, cid)
	result = decrypt(result, sck)
	return string(result[:])
}

func decrypt(str []byte, k string) []byte {
	var result []byte
	strls := len(str)
	strlk := len(k)
	for i := 0; i < strls; i++ {
		char := str[i]
		keychar := k[(i+strlk-1)%strlk]
		char = byte(((int(char) - int(keychar)) + 256) % 128)
		result = append(result, char)
	}
	return result
}

func reverse(s string) string {
	chars := []rune(s)
	for i, j := 0, len(chars)-1; i < j; i, j = i+1, j-1 {
		chars[i], chars[j] = chars[j], chars[i]
	}
	return string(chars)
}

func Hash(password string) ([]byte, error) {
	return bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
}

func VerifyHash(hashed, value string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashed), []byte(value))
}

var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func RandomString(n int) string {
	rand.Seed(time.Now().UnixNano())

	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	r := string(b)
	return r
}
