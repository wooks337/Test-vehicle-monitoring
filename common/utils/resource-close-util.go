package utils

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
)

func SetUpCloseHandler() {
	c := make(chan os.Signal, 2)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		fmt.Println("\r- Ctrl+C pressed in Termial")

		os.Exit(0)
	}()

}

var deferStack = []*func(){}

// ReleaseAllResource defer stack 에 저장된 모든 리소스를 해제한다.
// 최초 리소스를 등록하기 전에 defer 로 등록하기 위해 사용한다.
func ReleaseAllResource() {
	log.Println("Release All Resource")

	for i := 0; i < len(deferStack); i++ {
		/*
			defer로 등록하지 않고 호출만 하면 도중에 하나가 실패하면 중단 되므로
			defer 로 등록해서 defer stack의 모든 함수가 실행되게 한다.
		*/
		defer (*deferStack[i])()
	}
}

//  RegisterResourceCloser defer stack에 자원 해제 핸든러를 등록한다.
// defer RegisterResourceCloser() 호출 후 등록하면 된다.
func RegisterResourceCloser(closer func()) {
	fmt.Println("RegisterResourceCloser")
	deferStack = append(deferStack, &closer)
}
