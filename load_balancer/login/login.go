package login

import "load_balancer/engineering"

func UserLogin(userName string, passwd string) bool {
	redisPasswd, err := engineering.GetFromRedis(userName)
	if err != nil {
		return false
	}
	return passwd == redisPasswd
}
