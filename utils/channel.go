package utils

import (
	"regexp"
)

func GenerateChannelName(userName string, phone string) string {

	/*reg, _ := regexp.Compile("[^0-9]+")
	phone = reg.ReplaceAllString(phone, "")
	f := phone[0:4]
	l := phone[7:len(phone)]*/

	return userName + " " + phone
}

func MaskPhone(phone string) string {

	if len(phone) == 11 {
		reg, _ := regexp.Compile("[^0-9]+")
		phone = reg.ReplaceAllString(phone, "")
		f := phone[0:4]
		l := phone[7:len(phone)]
		return f + "***" + l
	}

	return phone

}
