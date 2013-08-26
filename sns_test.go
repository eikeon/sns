package sns_test

import (
	"log"
	"net/url"
	"sync"
	"testing"

	"github.com/eikeon/sns"
)

var SNS sns.SNS

func init() {
	SNS = sns.NewSNS()
}

func TestPublish(t *testing.T) {
	options := url.Values{"TopicArn": []string{"arn:aws:sns:us-east-1:966103638140:ginger-test"}}
	if _, err := SNS.Publish("Hello World", options); err != nil {
		t.Error(err)
	}
}

func bench() {
	options := url.Values{"TopicArn": []string{"arn:aws:sns:us-east-1:966103638140:ginger-test"}}
	if _, err := SNS.Publish("Hello World", options); err != nil {
		log.Println(err)
	}
}

func BenchmarkPublish(b *testing.B) {
	for i := 0; i < b.N; i++ {
		bench()
	}
}

func BenchmarkPublishConcurrent(b *testing.B) {
	C := 16

	items := make(chan int, b.N)
	for i := 0; i < b.N; i++ {
		items <- i
	}
	close(items)

	var wg sync.WaitGroup
	wg.Add(C)
	for i := 0; i < C; i++ {
		go func() {
			for _ = range items {
				bench()
			}
			wg.Done()
		}()
	}
	wg.Wait()

}
