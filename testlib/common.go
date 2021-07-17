package testlib

import (
	"github.com/saturninoabril/dashboard-server/model"
)

func GetTestEmail() string {
	return "user" + model.NewID() + "@example.com"
}
