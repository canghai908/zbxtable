package model

var (
	//IMWorkerChan   chan int
	WechatWorkerChan        chan int
	WechatRobotWorkerChan   chan int
	MailWorkerChan          chan int
	DingTalkRobotWorkerChan chan int
	FeishuRobotWorkerChan   chan int
	WebhookWorkerChan       chan int
)

func InitSenderWorker() {
	MailWorkerChan = make(chan int, 2)
	WechatWorkerChan = make(chan int, 2)
	WechatRobotWorkerChan = make(chan int, 2)
	DingTalkRobotWorkerChan = make(chan int, 2)
	FeishuRobotWorkerChan = make(chan int, 2)
	WebhookWorkerChan = make(chan int, 2)
}
