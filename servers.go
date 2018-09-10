package main

import (
	"runtime"
	"fmt"
	"net/http"
	"os"
)

var (
	MaxWorker = runtime.NumCPU()
	MaxQueue = 1000
)
type Serload struct{
	pri string
}
type Job struct {
	serload Serload
}
var JobQueue chan Job

type Worker struct {
	WorkerPool chan chan Job
	JobChannel chan Job
	Quit chan bool
}

func NewWorker(workPool chan chan Job)Worker  {
	return Worker{
		WorkerPool:workPool,
		JobChannel:make(chan Job),
		Quit:make(chan bool),
	}
}
func (w Worker) Start(){
	go func() {
		for {
			w.WorkerPool<-w.JobChannel
			select {
			case job :=<-w.JobChannel:
				fmt.Println(job.serload.pri)
			case <-w.Quit:
				return

			}
		}
	}()
}
func (w Worker) Stop(){
	go func(){
		w.Quit<-true
	}()
}

type Dispatcher struct{
	MaxWorkers int
	WorkerPool chan chan Job
	Quit chan bool
}
func  NewDispatcher(maxWorkers int ) *Dispatcher{//新调度器
	pool :=make(chan chan Job,maxWorkers)
	return &Dispatcher{maxWorkers,pool,make(chan bool)}
}
func (d *Dispatcher)Run(){
	for i:=0;i<d.MaxWorkers;i++{
		worker :=NewWorker(d.WorkerPool)
		worker.Start()
	}
	go d.Dispatch()
}
func (d *Dispatcher) Stop(){
	go func() {
		d.Quit<-true
	}()
}
func (d *Dispatcher)Dispatch()  {//调度器
	for {
		select {
		case job:=<- JobQueue:
			go func(job Job) {
				jobChannel :=<-d.WorkerPool
				jobChannel<-job
			}(job)
		case<-d.Quit:
			return
		}
	}
}

func entry(res http.ResponseWriter,req *http.Request)  {//获取到消息的条目
	work :=Job{serload:Serload{pri:"Ok,let's go!"}}
	JobQueue<-work
	fmt.Println(res,"Hello World again")
}

func init(){//初始化
	runtime.GOMAXPROCS(MaxWorker)
	JobQueue=make(chan Job,MaxQueue)
	dispatcher:=NewDispatcher(MaxWorker)
	dispatcher.Run()
}

func main()  {
	Port:="8080"
	IsHttp :=true
	arg_num :=len(os.Args)
	if 2<=arg_num{
		Port = os.Args[1]
	}
	if 3<=arg_num{
		IsHttp= true
	}else {
		IsHttp = false
	}
	fmt.Println("server is http %t\n",IsHttp)
	fmt.Println("server listens at",Port)
	http.HandleFunc("/",entry)

	var err error
	if IsHttp{
		err = http.ListenAndServe(":"+Port,nil)
	}else {
		err = http.ListenAndServe(":"+Port,nil)
	}
	if err !=nil{
		fmt.Println("server failure",err)
	}
	fmt.Println("quit")

}
