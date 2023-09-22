package login

import "load_balancer/redisDB"

func UserLogin(userName string, passwd string) bool {
	redisPasswd, err := redisDB.GetFromRedis(userName)
	if err != nil {
		return false
	}
	return passwd == redisPasswd
}
