package helpers

import (
	"encoding/json"
	"math/big"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"unsafe"

	"github.com/google/uuid"
	"github.com/leekchan/accounting"
	"github.com/tidwall/sjson"
	"gitlab.com/pangestu18/janji-online/chat/helpers/phpserialize"
)

type Iconvert struct {
	Val interface{}
}

type UnserializedPhpSession struct {
	Val phpserialize.PhpValue
}

var link = regexp.MustCompile("(^[A-Za-z])|_([A-Za-z])")
var matchFirstCap = regexp.MustCompile("(.)([A-Z][a-z]+)")
var matchAllCap = regexp.MustCompile("([a-z0-9])([A-Z])")

func ToCamelCase(str string) string {
	return link.ReplaceAllStringFunc(str, func(s string) string {
		return strings.ToUpper(strings.Replace(s, "_", "", -1))
	})
}

func ToSnakeCase(str string) string {
	snake := matchFirstCap.ReplaceAllString(str, "${1}_${2}")
	snake = matchAllCap.ReplaceAllString(snake, "${1}_${2}")
	return strings.ToLower(snake)
}

func Float64ToCurrency(n float64) string {
	ac := accounting.Accounting{Symbol: "Rp ", Thousand: ".", Decimal: ",", Precision: 0}
	return ac.FormatMoneyBigFloat(big.NewFloat(n))
}

func DotToInterface(data []map[string]interface{}, schema map[string]interface{}) []map[string]interface{} {
	temp := `[]`
	i := -1
	for _, d := range data {
		i++
		for k, v := range d {
			for f, c := range schema["fields"].(map[string]map[string]string) {
				if strings.ToLower(k) == c["as"] && v != nil {
					if c["type"] == "float64" {
						val, _ := strconv.ParseFloat(string(v.([]byte)), 64)
						temp = DotNotationSet(temp, strconv.Itoa(i)+"."+f, val)
					} else if c["type"] == "int64" {
						temp = DotNotationSet(temp, strconv.Itoa(i)+"."+f, v.(int64))
					} else if c["type"] == "boolean" {
						switch v.(type) {
						case string:
							vbol := v == "true" || v == "T"
							temp = DotNotationSet(temp, strconv.Itoa(i)+"."+f, vbol)
						case int, int64, uint, uint64, int8, uint8:
							vbol := v.(int64) == 1
							temp = DotNotationSet(temp, strconv.Itoa(i)+"."+f, vbol)
						}
					} else if c["type"] == "pq.StringArray" {
						val := string(v.([]byte))
						val = strings.ReplaceAll(val, "{", "")
						val = strings.ReplaceAll(val, "}", "")
						temp = DotNotationSet(temp, strconv.Itoa(i)+"."+f, strings.Split(val, ","))
					} else if v != "" {
						temp = DotNotationSet(temp, strconv.Itoa(i)+"."+f, v)
					}
				}
			}
		}
	}

	var res []map[string]interface{}
	json.Unmarshal([]byte(temp), &res)
	return res
}

func DotNotationSet(json, path string, value interface{}) string {
	res, _ := sjson.Set(json, path, value)
	return res
}

func NewUUID() string {
	return uuid.New().String()
}

func ValidateUUID(id string) (bool, string) {
	res, err := uuid.Parse(id)
	if err != nil {
		return false, ""
	} else {
		return true, res.String()
	}
}

func NewToken() string {
	return strings.ReplaceAll(NewUUID(), "-", "")
}

func Convert(val interface{}) Iconvert {
	return Iconvert{Val: val}
}

func (v Iconvert) String() string {
	switch v.Val.(type) {
	case string:
		return v.Val.(string)
	case bool:
		return strconv.FormatBool(v.Val.(bool))
	case float32:
		return strconv.FormatFloat(float64(v.Val.(float32)), 'E', -1, 32)
	case float64:
		return strconv.FormatFloat(v.Val.(float64), 'E', -1, 64)
	case uint:
		return strconv.FormatUint(uint64(v.Val.(uint)), 10)
	case uint8:
		return strconv.FormatUint(uint64(v.Val.(uint8)), 10)
	case uint16:
		return strconv.FormatUint(uint64(v.Val.(uint16)), 10)
	case uint32:
		return strconv.FormatUint(uint64(v.Val.(uint32)), 10)
	case uint64:
		return strconv.FormatUint(v.Val.(uint64), 10)
	case uintptr:
		return strconv.FormatUint(uint64(v.Val.(uintptr)), 10)
	case int:
		return strconv.FormatInt(int64(v.Val.(int)), 10)
	case int8:
		return strconv.FormatInt(int64(v.Val.(int8)), 10)
	case int16:
		return strconv.FormatInt(int64(v.Val.(int16)), 10)
	case int32:
		return strconv.FormatInt(int64(v.Val.(int32)), 10)
	case int64:
		return strconv.FormatInt(v.Val.(int64), 10)
	default:
		return ""
	}
}

func (v Iconvert) Int() int {
	val, err := strconv.Atoi(v.String())
	if err != nil {
		return 0
	}
	return val
}

func (v Iconvert) Bool() bool {
	b := false
	switch v.Val.(type) {
	case string:
		if v.Val == "1" || v.Val == "T" {
			b = true
		}
	}
	return b
}

func (v Iconvert) ConvertTo(t string) interface{} {
	if t == "boolean" {
		return v.Bool()
	} else if t == "integer" {
		return v.Int()
	} else {
		return v.String()
	}
}

func PhpSerialize(v interface{}) (string, error) {
	encoder := phpserialize.NewSerializer()
	switch v.(type) {
	case map[string]interface{}:
		var source phpserialize.PhpArray = make(phpserialize.PhpArray)
		for i, k := range v.(map[string]interface{}) {
			source[i] = k
		}
		if r, err := encoder.Encode(source); err != nil {
			return "", err
		} else {
			return r, nil
		}
	default:
		source := v
		if r, err := encoder.Encode(source); err != nil {
			return "", err
		} else {
			return r, nil
		}
	}
}

func PhpUnserialize(v string) (UnserializedPhpSession, error) {
	decoder := phpserialize.NewUnSerializer(v)
	if r, err := decoder.Decode(); err != nil {
		return UnserializedPhpSession{}, err
	} else {
		return UnserializedPhpSession{Val: r}, nil
	}
}

func (v UnserializedPhpSession) String() string {
	return phpserialize.PhpValueString(v.Val)
}

func ReplaceAtIndex(in string, r rune, i int) string {
	out := []rune(in)
	out[i] = r
	return string(out)
}

func (v UnserializedPhpSession) Map() map[interface{}]interface{} {
	result := make(map[interface{}]interface{})
	switch v.Val.(type) {
	case phpserialize.PhpArray:
		for i, x := range v.Val.(phpserialize.PhpArray) {
			result[i] = x
		}
	}
	return result
}

func (v UnserializedPhpSession) GetObject() *phpserialize.PhpObject {
	return v.Val.(*phpserialize.PhpObject)
}

func (v UnserializedPhpSession) GetFromMap(k interface{}) interface{} {
	m := v.Map()
	return m[k]
}

func BytesToString(b []byte) string {
	bh := (*reflect.SliceHeader)(unsafe.Pointer(&b))
	sh := reflect.StringHeader{bh.Data, bh.Len}
	return *(*string)(unsafe.Pointer(&sh))
}

func GetMinimum(a, b int64) int64 {
	if a < b {
		return a
	}
	return b
}
