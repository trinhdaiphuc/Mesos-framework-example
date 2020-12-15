package scheduler

import (
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"sync"

	"github.com/carlonelong/mesos-framework-sdk/client"
	mesos_v1 "github.com/carlonelong/mesos-framework-sdk/include/mesos/v1"
	mesos_v1_scheduler "github.com/carlonelong/mesos-framework-sdk/include/mesos/v1/scheduler"
	"github.com/carlonelong/mesos-framework-sdk/logging"
	"github.com/carlonelong/mesos-framework-sdk/resources"
	"github.com/carlonelong/mesos-framework-sdk/scheduler"
	uuid "github.com/satori/go.uuid"
	"github.com/trinhdaiphuc/Mesos-framework-example/config"
)

type Scheduler struct {
	FrameworkInfo  *mesos_v1.FrameworkInfo
	CommandInfoURI []*mesos_v1.CommandInfo_URI
	Client         client.Client
	TaskLaunch     int
	sync.Mutex
}

func toJsonString(t interface{}) string {
	result, _ := json.Marshal(t)
	return string(result)
}

func NewScheduler(cfg config.Config, executorURI string) *Scheduler {
	logger := logging.NewDefaultLogger()
	frameworkInfo := &mesos_v1.FrameworkInfo{
		User:            config.ProtoString(cfg.User),
		Name:            config.ProtoString(cfg.Name),
		FailoverTimeout: config.ProtoFloat64(cfg.FailoverTimeout.Seconds()),
		Checkpoint:      config.ProtoBool(cfg.Checkpoint),
		Role:            config.ProtoString(cfg.Role),
		Hostname:        config.ProtoString(cfg.Hostname),
	}

	// Create http client to handle calls and events framework
	clientScheduler := client.NewClient(
		client.ClientData{
			Endpoint: cfg.SchedulerEndpoint,
		}, logger)

	commandURI := []*mesos_v1.CommandInfo_URI{
		{
			Value:      config.ProtoString(executorURI),
			Executable: config.ProtoBool(true),
		},
	}
	return &Scheduler{
		FrameworkInfo:  frameworkInfo,
		CommandInfoURI: commandURI,
		Client:         clientScheduler,
		TaskLaunch:     0,
	}
}

func (s *Scheduler) Run(logger logging.Logger) {
	schedEvent := make(chan *mesos_v1_scheduler.Event)
	sched := scheduler.NewDefaultScheduler(s.Client, s.FrameworkInfo, logger)

	go func() {
		_, err := sched.Subscribe(schedEvent)
		if err != nil {
			log.Fatal("Error scheduler ", err)
		}
	}()

	for {
		event := <-schedEvent
		switch event.Type.String() {
		case mesos_v1_scheduler.Event_SUBSCRIBED.String():
			fmt.Printf("Scheduler subscribe %+v\n", toJsonString(event.Subscribed))
			s.FrameworkInfo.Id = event.Subscribed.FrameworkId
		case mesos_v1_scheduler.Event_OFFERS.String():
			fmt.Printf("Scheduler offers %+v\n", toJsonString(event.Offers))
			go s.Accept(sched, event.Offers.Offers)
		case mesos_v1_scheduler.Event_MESSAGE.String():
			fmt.Println("Executor message", string(event.Message.Data))
		}
	}

}

func createResources(offer *mesos_v1.Offer, cpu, mem float64) []*mesos_v1.Resource {
	cpuRes := resources.CreateResource("cpus", "*", cpu)
	cpuRes.AllocationInfo = offer.AllocationInfo
	memRes := resources.CreateResource("mem", "*", mem)
	memRes.AllocationInfo = offer.AllocationInfo
	return []*mesos_v1.Resource{cpuRes, memRes}
}

func createTaskInfo(agentID *mesos_v1.AgentID, executorInfo *mesos_v1.ExecutorInfo, taskLaunch int, resource []*mesos_v1.Resource) *mesos_v1.TaskInfo {
	taskID := &mesos_v1.TaskID{Value: config.ProtoString(strconv.Itoa(taskLaunch))}
	return &mesos_v1.TaskInfo{
		Name:      config.ProtoString("Test-Task-" + *taskID.Value),
		TaskId:    taskID,
		AgentId:   agentID,
		Resources: resource,
		Executor:  executorInfo,
	}
}

func (s *Scheduler) Accept(sched *scheduler.DefaultScheduler, offers []*mesos_v1.Offer) {
	s.Lock()
	s.TaskLaunch++
	if s.TaskLaunch > 5 {
		s.Unlock()
		return
	}
	s.Unlock()

	// Create ressources
	res := createResources(offers[0], 1.0, 128.0)

	// Create executor info
	executor := &mesos_v1.ExecutorInfo{
		Type:        mesos_v1.ExecutorInfo_CUSTOM.Enum(),
		ExecutorId:  &mesos_v1.ExecutorID{Value: config.ProtoString(uuid.NewV1().String())},
		FrameworkId: s.FrameworkInfo.Id,
		Command:     resources.CreateSimpleCommandInfo(config.ProtoString("./executor"), s.CommandInfoURI),
		Name:        config.ProtoString("Test-Executor"),
	}

	offerIDs := []*mesos_v1.OfferID{offers[0].Id}

	// Create task info
	task := createTaskInfo(offers[0].AgentId, executor, s.TaskLaunch, res)

	// Create operation
	operation := &mesos_v1.Offer_Operation{
		Type:   mesos_v1.Offer_Operation_LAUNCH.Enum(),
		Launch: &mesos_v1.Offer_Operation_Launch{TaskInfos: []*mesos_v1.TaskInfo{task}},
	}
	operations := []*mesos_v1.Offer_Operation{operation}

	filter := &mesos_v1.Filters{RefuseSeconds: config.ProtoFloat64(5.0)}

	_, err := sched.Accept(offerIDs, operations, filter)
	if err != nil {
		fmt.Println("Scheduler accept event error", err)
	}
}
