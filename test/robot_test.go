package test

import (
	"fmt"
	"rk-api/pkg/logger"
	"testing"
)

func InitLogger() (func(), error) {
	logger := logger.GetLogger()
	defer logger.Sync()

	return func() {}, nil
}

// http://localhost:8080/public/robot.html
// go test -v -run TestRobt
func TestRobt(t *testing.T) {

	// t.Log("hello")
	InitLogger()

	accounts, passwords, err := ReadCredentials()
	if err != nil {
		fmt.Println("Error reading credentials:", err)
	}

	// fmt.Println("Accounts and passwords:")
	for i := range accounts {
		fmt.Printf("%s: %s\n", accounts[i], passwords[i])
		robot := NewApiRobot(accounts[i], passwords[i])
		go robot.Start()

		if i > 3 {
			break
		}
	}

	// BatchGenAccount("190500600", 889, 2000)

	select {}

}

func BatchGenAccount(mobile string, startAt, count int) {
	for i := startAt; i < count; i++ {
		account := fmt.Sprintf("%s%d", mobile, i)
		password := "123456"
		robot := NewApiRobot(account, password)
		robot.register()
	}
}

// go func() {
// 	rot, err := telegram.InitBot()
// 	t.Log(rot, err)
// }()

// readyToProcess := make(chan struct{}, 1) // 非阻塞通道
// // readyToProcess <- struct{}{}
// select {
// case readyToProcess <- struct{}{}:
// default:
// }
