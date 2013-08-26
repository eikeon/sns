package sns

import (
	"encoding/xml"
	"errors"
	"io"
	"io/ioutil"
	"log"
	"math"
	"net/http"
	"net/url"
	"time"

	"github.com/eikeon/aws4"
)

type PublishResponse struct {
	MessageId string `xml:"PublishResult>MessageId"`
}

type SNS interface {
	Publish(message string, options url.Values) (*PublishResponse, error)
}

type sns struct {
	client *aws4.Client
}

func NewSNS() SNS {
	d := &sns{}
	if d.getClient() == nil {
		log.Println("could not create sns: no default aws4 client")
		return nil
	}
	return d
}

func (s *sns) getClient() *aws4.Client {
	if s.client == nil {
		s.client = aws4.DefaultClient
		if s.client != nil {
			tr := &http.Transport{MaxIdleConnsPerHost: 100}
			c := &http.Client{Transport: tr}
			s.client.Client = c
		}

	}
	return s.client
}

func (s *sns) post(action string, data url.Values) (io.ReadCloser, error) {
	url := "https://sns.us-east-1.amazonaws.com/"
	currentRetry := 0
	maxNumberOfRetries := 10
	data["Action"] = []string{action}
RETRY:

	if response, err := s.getClient().PostForm(url, data); err == nil {
		switch response.StatusCode {
		case 200:
			return response.Body, nil
		case 500:
			response.Body.Close()
			log.Println("Got a 500 error... retrying.")

		default:
			b, err := ioutil.ReadAll(response.Body)
			response.Body.Close()
			if err != nil {
				return nil, err
			}
			return nil, errors.New(string(b))
		}
	} else {
		return nil, err
	}
	if currentRetry < maxNumberOfRetries {
		wait := time.Duration(math.Pow(2, float64(currentRetry))) * 50 * time.Millisecond
		time.Sleep(wait)
		currentRetry = currentRetry + 1
		goto RETRY
	} else {
		return nil, errors.New("exceeded maximum number of retries")
	}
}

func (s *sns) Publish(message string, options url.Values) (*PublishResponse, error) {
	options["Message"] = []string{message}
	if reader, err := s.post("Publish", options); err == nil {
		response := &PublishResponse{}
		if err = xml.NewDecoder(reader).Decode(&response); err != nil {
			return nil, err
		}
		reader.Close()
		return response, nil
	} else {
		return nil, err
	}
}
