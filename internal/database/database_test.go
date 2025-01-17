package database

import (
	"context"
	"fmt"
	"log"
	"testing"

	"github.com/testcontainers/testcontainers-go/modules/redis"
)

func mustStartRedisContainer() (func(context.Context) error, error) {
	dbContainer, err := redis.Run(
		context.Background(),
		"docker.io/redis:7.2.4",
		redis.WithSnapshotting(10, 1),
		redis.WithLogLevel(redis.LogLevelVerbose),
	)
	if err != nil {
		return nil, err
	}

	dbHost, err := dbContainer.Host(context.Background())
	if err != nil {
		return dbContainer.Terminate, err
	}

	dbPort, err := dbContainer.MappedPort(context.Background(), "6379/tcp")
	if err != nil {
		return dbContainer.Terminate, err
	}

	address = dbHost
	port = dbPort.Port()
	database = "0"

	return dbContainer.Terminate, err
}

func TestGetSetValueRedis(t *testing.T) {
	srv := New()
	srv.SetValueRedis(context.Background(), "test", 0)
	str, err := srv.GetValueRedis(context.Background(), "test")
	if err != nil {
		log.Fatalf("failed to get test: %v", err)
	}
	log.Printf("answer: %v", str)
	intTest, popErr := srv.PopValueRedis(context.Background(), "test")
	if popErr != nil {
		log.Fatalf("failed to get test: %v", err)
	}
	log.Printf("answer: %v", intTest)
	str, err = srv.GetValueRedis(context.Background(), "test")
	if err == nil {
		log.Fatalf("failed to get test: %v", err)
	}
	log.Printf("successfully failed to find deleted item %v", str)
}

func TestMain(m *testing.M) {
	teardown, err := mustStartRedisContainer()
	if err != nil {
		log.Fatalf("could not start redis container: %v", err)
	}

	m.Run()

	if teardown != nil && teardown(context.Background()) != nil {
		log.Fatalf("could not teardown redis container: %v", err)
	}
}

func TestNew(t *testing.T) {
	srv := New()
	if srv == nil {
		t.Fatal("New() returned nil")
	}
}

func TestHealth(t *testing.T) {
	srv := New()

	stats := srv.Health()

	if stats["redis_status"] != "up" {
		t.Fatalf("expected status to be up, got %s", stats["redis_status"])
	}

	if _, ok := stats["redis_version"]; !ok {
		t.Fatalf("expected redis_version to be present, got %v", stats["redis_version"])
	}
}

func TestConnectToDB(t *testing.T) {
	res, err := ConnectToDB()

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if res == nil {
		t.Fatalf("expected true, got %v", res)
	}
	res.Close()
}

func TestInsertInto(t *testing.T) {
	contents := []string{"test", "this is a test", "fake path", "test"}
	columns := []string{"name", "description", "path", "quality"}
	err := InsertInto("videos", columns, contents)

	if err != nil {
		t.Fatalf("insert into expected no error, got %v", err)
	}
	res, geterr := GetAllFromTable("videos")

	fmt.Println(res, geterr)

	if geterr != nil {
		t.Fatalf("get all from table expected no error, got %v", err)
	}
	if len(res) <= 0 {
		t.Fatalf("expected more than 1, got %v", len(res))
	}
	if res[0]["name"] != "test" {
		t.Fatalf("expected test, got %v", res[0])
	}
}

func TestGetAllFromTable(t *testing.T) {
	_, err := GetAllFromTable("videos")

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}
