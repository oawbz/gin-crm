package common

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"strconv"
)

func GetMD5Hash(text string) string {
	//md5字符串
	hash := md5.Sum([]byte(text))
	return hex.EncodeToString(hash[:])
}

func Txt(s interface{}) string {
	//类型转换
	return fmt.Sprintf("%v", s)
}

func ToInt(s interface{}) int64 {
	//类型转换
	//int64, _ := strconv.Atoi(txt(s))

	int64, err := strconv.ParseInt(Txt(s), 10, 64)
	if err != nil {
		return 0
	}
	return int64
}

/*
func RemoveDuplicateElement(languages []string) []string { //map去重
	result := make([]string, 0, len(languages))
	temp := map[string]struct{}{}
	for _, item := range languages {
		if _, ok := temp[item]; !ok {
			temp[item] = struct{}{}
			result = append(result, item)
		}
	}
	return result
}
*/
