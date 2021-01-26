# Mesos golang framework demo

The demo shows how to implement a custom Executor and Scheduler communicating with Mesos masters and slaves. For easy and fast development, I use [Mesos framework SDK](https://github.com/carlonelong/mesos-framework-sdk) to create a demo that executes 5 simple tasks through HTTP API wrapped for Calls and Events handlers. 

### How to run the demo framework:
- Assume there is Mesos master endpoint is running on http://127.0.0.1:5050
- Clone repository: 

```
git clone https://github.com/trinhdaiphuc/Mesos-framework-example
```

- Set up environments: the demo support .env file for an easier setup like the example above or `.env.example` file:

```sh
# Framework user to register with the Mesos master
FRAMEWORK_USER=root
# Framework nname to register with the Mesos master
FRAMEWORK_NAME=example
# Framework role to register with the Mesos master
FRAMEWORK_ROLE=*
# The endpoint of Mesos is running on
MESOS_ENDPOINT=127.0.0.1:5050
# The path to executor binary file
EXEC_BINARY=./bin/executor
# The server for serving executor binary file
EXEC_SERVER_PORT=5656
# Framework failover timeout (recover from scheduler failure)
SCHEDULER_FAILOVER_TIMEOUT=1m
# Mesos scheduler API connection timeout
MESOS_CONNECT_TIMEOUT=20s
# Framework hostname that is advertised to the master
HOST=localhost
```

Build Executor: pay attention to the executor path output. The example shows that the executor binary output path is ./bin/executor (after -o flag in command go build)

```go
go build -o bin/executor ./executor
```

Run Scheduler:

```go
go run main.go
```

It can be run script in Makefile
```sh
make run
``` 
