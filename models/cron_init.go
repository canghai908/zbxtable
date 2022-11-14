package models

var (
	//IMWorkerChan   chan int
	WechatWorkerChan chan int
	MailWorkerChan   chan int
)

func InitSenderWorker() {
	MailWorkerChan = make(chan int, 2)
	WechatWorkerChan = make(chan int, 2)
}
