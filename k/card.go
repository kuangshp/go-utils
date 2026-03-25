package k

import (
	"strconv"
	"time"
)

// GetBirthFromIDCard 从身份证号获取出生日期
func GetBirthFromIDCard(idCard string) string {
	if len(idCard) != 18 {
		return ""
	}
	return idCard[6:10] + "-" + idCard[10:12] + "-" + idCard[12:14]
}

// GetAgeFromIDCard 从身份证号获取年龄
func GetAgeFromIDCard(idCard string) int {
	if len(idCard) != 18 {
		return 0
	}
	birthYear, _ := strconv.Atoi(idCard[6:10])
	currentYear := time.Now().Year()
	return currentYear - birthYear
}

// GetGenderFromIDCard 从身份证号获取性别
func GetGenderFromIDCard(idCard string) string {
	if len(idCard) != 18 {
		return ""
	}
	genderCode, _ := strconv.Atoi(idCard[16:17])
	if genderCode%2 == 0 {
		return "女"
	}
	return "男"
}
