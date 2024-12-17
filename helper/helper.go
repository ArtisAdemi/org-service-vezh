package helper

import "math/rand"


var letterRunes = []rune("123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func RandomString(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}

func RoleTypeStr(roleId int) string {
	switch roleId {
	case 1:
		return "Owner"
	case 2:
		return "Admin"
	case 3:
		return "Coach"
	case 4:
		return "SME"
	case 5:
		return "Client Alumn"
	case 6:
		return "Client Current"
	case 7:
		return "Client Future"
	case 8:
		return "Partner"
	case 9:
		return "Guest"
	}
	return ""
}