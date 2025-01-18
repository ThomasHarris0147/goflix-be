# Project goflix-be

Hello, this is a simple test project where I wanted to see if I could serve videos using kafka task queues and redis as a local cache. I left some empty files to indicate where cloud services could be integrated but I was not interested in paying a few dollars for a test cloud server yet.


![goflix-be drawio (2)](https://github.com/user-attachments/assets/9af94e4a-e7c8-4712-a752-cbc94c992603)


Side note: the GET /videos with Optional Name and Quality is still WIP but right now you can retrieve videos Name and Quality (although currently required).

## Getting Started

These instructions will get you a copy of the project up and running on your local machine for development and testing purposes. See deployment for notes on how to deploy the project on a live system.

## MakeFile

Run build make command with tests
```bash
make all
```

Build the application
```bash
make build
```

Run the application
```bash
make run
```
Create DB container
```bash
make docker-run
```

Shutdown DB Container
```bash
make docker-down
```

DB Integrations Test:
```bash
make itest
```

Live reload the application:
```bash
make watch
```

Run the test suite:
```bash
make test
```

Clean up binary from the last build:
```bash
make clean
```
