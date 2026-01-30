package model

var (
	//IMWorkerChan   chan int
	WechatWorkerChan      chan int
	WechatRobotWorkerChan chan int
	MailWorkerChan        chan int
)

func InitSenderWorker() {
	MailWorkerChan = make(chan int, 2)
	WechatWorkerChan = make(chan int, 2)
	WechatRobotWorkerChan = make(chan int, 2)
}
