package main

import (
	"fmt"
	"log"
	"net/url"
	"os"
	"time"

	"github.com/carlonelong/mesos-framework-sdk/client"
	"github.com/carlonelong/mesos-framework-sdk/executor"
	mesos_v1 "github.com/carlonelong/mesos-framework-sdk/include/mesos/v1"
	mesos_v1_executor "github.com/carlonelong/mesos-framework-sdk/include/mesos/v1/executor"
	"github.com/carlonelong/mesos-framework-sdk/logging"
	"github.com/mesos/mesos-go/api/v1/lib/executor/config"
	uuid "github.com/satori/go.uuid"
)

var apiPath = "/api/v1/executor"

func Run(client client.Client, logger logging.Logger, cfg config.Config) {
	frameworkID := &mesos_v1.FrameworkID{Value: protoString(cfg.FrameworkID)}
	executorID := &mesos_v1.ExecutorID{Value: protoString(cfg.ExecutorID)}

	exec := executor.NewDefaultExecutor(frameworkID, executorID, client, logger)
	execEvent := make(chan *mesos_v1_executor.Event)

	go func() {
		if err := exec.Subscribe(execEvent); err != nil {
			fmt.Println("Error executor ", err)
		}
	}()

	for {
		event := <-execEvent
		switch event.Type.String() {
		case mesos_v1_executor.Event_SUBSCRIBED.String():
			fmt.Println("Executor subscribed")
			go exec.Message([]byte("Executor subscribed"))
		case mesos_v1_executor.Event_LAUNCH.String():
			fmt.Println("Executor event launch")
			go update(exec, event.Launch)
		case mesos_v1_executor.Event_SHUTDOWN.String():
			fmt.Println("Executor event shutdown")
			return
		}
	}
}

// update task status
func update(exec executor.Executor, event *mesos_v1_executor.Event_Launch) {
	// Update task status to RUNNING
	err := exec.Update(createTaskUpdate(event.Task.TaskId, mesos_v1.TaskState_TASK_RUNNING))
	if err != nil {
		exec.Update(createTaskUpdate(event.Task.TaskId, mesos_v1.TaskState_TASK_FAILED))
		fmt.Printf("Task failed when running: TaskID=%v, TaskUUID=%v error: %v\n", event.Task.TaskId.String(), err)
		exec.Message([]byte(err.Error()))
		return
	} else {
		fmt.Println("Task is running")
	}

	// TODO: Do some logics here
	time.Sleep(3 * time.Second)

	// After finish logic code update task status to FINISHED
	err = exec.Update(createTaskUpdate(event.Task.TaskId, mesos_v1.TaskState_TASK_FINISHED))
	if err != nil {
		exec.Update(createTaskUpdate(event.Task.TaskId, mesos_v1.TaskState_TASK_FAILED))
		fmt.Printf("Task failed when finnish: TaskID=%v, error: %v\n", event.Task.TaskId.String(), err)
		exec.Message([]byte(err.Error()))
		return
	} else {
		fmt.Println("Task finished")
	}
}

func createTaskUpdate(taskID *mesos_v1.TaskID, state mesos_v1.TaskState) *mesos_v1.TaskStatus {
	return &mesos_v1.TaskStatus{
		TaskId:    taskID,
		State:     state.Enum(),
		Source:    mesos_v1.TaskStatus_SOURCE_EXECUTOR.Enum(),
		Timestamp: protoFloat64(float64(time.Now().Unix())),
		Uuid:      uuid.NewV4().Bytes(),
	}
}

func main() {
	// Get environment variables from Mesos slave
	cfg, err := config.FromEnv()
	if err != nil {
		log.Fatal("failed to load configuration: " + err.Error())
	}

	// Create API endpoint for executor
	apiURL := url.URL{
		Scheme: "http",
		Host:   cfg.AgentEndpoint,
		Path:   apiPath,
	}
	logger := logging.NewDefaultLogger()

	// Create http client to handle calls and events framework
	clientExecutor := client.NewClient(client.ClientData{Endpoint: apiURL.String()}, logger)

	// Run executor
	fmt.Println("Executor is running...")
	Run(clientExecutor, logger, cfg)
	fmt.Println("Executor is stopped")
	os.Exit(0)
}

func protoString(s string) *string    { return &s }
func protoFloat64(f float64) *float64 { return &f }
