package cfg

import "os"

var JwtSecret []byte

func Init() {
	jwtSecret, err := os.ReadFile("/run/secrets/jwt_secret")
	if err != nil {
		panic(err.Error())
	}
	JwtSecret = jwtSecret
}
