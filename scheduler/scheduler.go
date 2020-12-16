package scheduler

import (
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"strconv"
	"sync"

	"github.com/carlonelong/mesos-framework-sdk/client"
	mesos_v1 "github.com/carlonelong/mesos-framework-sdk/include/mesos/v1"
	mesos_v1_scheduler "github.com/carlonelong/mesos-framework-sdk/include/mesos/v1/scheduler"
	"github.com/carlonelong/mesos-framework-sdk/logging"
	"github.com/carlonelong/mesos-framework-sdk/resources"
	"github.com/carlonelong/mesos-framework-sdk/scheduler"
	"github.com/carlonelong/mesos-framework-sdk/utils"
	uuid "github.com/satori/go.uuid"
	"github.com/trinhdaiphuc/Mesos-framework-example/config"
)

type Scheduler struct {
	FrameworkInfo  *mesos_v1.FrameworkInfo
	CommandInfoURI []*mesos_v1.CommandInfo_URI
	Client         client.Client
	TaskLaunch     int // The counter task number
	sync.Mutex
}

var apiPath = "/api/v1/scheduler"

func toJsonString(t interface{}) string {
	result, _ := json.Marshal(t)
	return string(result)
}

func NewScheduler(cfg config.Config, executorURI string) *Scheduler {
	logger := logging.NewDefaultLogger()
	frameworkInfo := &mesos_v1.FrameworkInfo{
		User:            utils.ProtoString(cfg.User),
		Name:            utils.ProtoString(cfg.Name),
		FailoverTimeout: utils.ProtoFloat64(cfg.FailoverTimeout.Seconds()),
		Checkpoint:      utils.ProtoBool(cfg.Checkpoint),
		Role:            utils.ProtoString(cfg.Role),
		Hostname:        utils.ProtoString(cfg.Hostname),
	}
	// Create API endpoint for executor
	apiURL := url.URL{
		Scheme: "http",
		Host:   cfg.MesosEndpoint,
		Path:   apiPath,
	}

	// Create http client to handle calls and events framework
	clientScheduler := client.NewClient(client.ClientData{Endpoint: apiURL.String()}, logger)

	// Set up executor command from executor endpoint
	commandURI := []*mesos_v1.CommandInfo_URI{
		{
			Value:      utils.ProtoString(executorURI),
			Executable: utils.ProtoBool(true),
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
			// Create tasks and send Accept request to tell Executor launch these
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
	taskID := &mesos_v1.TaskID{Value: utils.ProtoString(strconv.Itoa(taskLaunch))}
	return &mesos_v1.TaskInfo{
		Name:      utils.ProtoString("Test-Task-" + *taskID.Value),
		TaskId:    taskID,
		AgentId:   agentID,
		Resources: resource,
		Executor:  executorInfo,
	}
}

// checkTaskLaunch return true if TaskLaunch counter is not greater than total
func (s *Scheduler) checkTaskLaunch(total int) bool {
	s.Lock()
	defer s.Unlock()
	if s.TaskLaunch >= total {
		return false
	}
	s.TaskLaunch++
	return true
}

func (s *Scheduler) Accept(sched *scheduler.DefaultScheduler, offers []*mesos_v1.Offer) {
	// Increase counter and create only 5 tasks
	if !s.checkTaskLaunch(5) {
		return
	}

	// Create ressources
	res := createResources(offers[0], 1.0, 128.0)

	// Create executor info
	executor := &mesos_v1.ExecutorInfo{
		Type:        mesos_v1.ExecutorInfo_CUSTOM.Enum(),
		ExecutorId:  &mesos_v1.ExecutorID{Value: utils.ProtoString(uuid.NewV1().String())},
		FrameworkId: s.FrameworkInfo.Id,
		Command:     resources.CreateSimpleCommandInfo(utils.ProtoString("./executor"), s.CommandInfoURI),
		Name:        utils.ProtoString("Test-Executor"),
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

	filter := &mesos_v1.Filters{RefuseSeconds: utils.ProtoFloat64(5.0)}

	_, err := sched.Accept(offerIDs, operations, filter)
	if err != nil {
		fmt.Println("Scheduler accept event error", err)
	}
}
