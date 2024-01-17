package tests

import (
	"github.com/davecgh/go-spew/spew"
	"github.com/stretchr/testify/assert"
	"github.com/yaklang/yaklang/common/log"
	"github.com/yaklang/yaklang/common/mq"
	"github.com/yaklang/yaklang/common/thirdpartyservices"
	"github.com/yaklang/yaklang/common/utils"
	"io/ioutil"
	"strings"
	"testing"
	"time"
)

func Test_Conn(t *testing.T) {
	//Obtain the ampq address
	u := thirdpartyservices.GetAMQPUrl()

	test := assert.New(t)
	//Set the message broker, the second parameter is the config method to override the default value
	broker, err := mq.NewBroker(utils.TimeoutContext(5*time.Second), mq.WithAMQPUrl(u))
	if err != nil {
		test.FailNow(err.Error())
	}
	//Create a broker server, set the address to test?? Concurrency 200
	server, err := mq.NewListener(broker, "test", 200)
	if err != nil {
		test.FailNow(err.Error())
		return
	}
	err = broker.RunBackground()
	if err != nil {
		test.FailNow(err.Error())
		return
	}

	log.Infof("server is started")
	client, err := mq.NewConnectionWithBroker("test-client", "test", broker)
	if err != nil {
		test.FailNow(err.Error())
		return
	}

	log.Info("client is created")

	go func() {
		client.Write([]byte("hello"))
		time.Sleep(2 * time.Second)
		client.Close()
	}()

	time.Sleep(1 * time.Second)
	conn, err := server.Accept()
	if err != nil {
		test.FailNow(err.Error())
		return
	}

	log.Info("start to recv data from client")
	bytes, _ := ioutil.ReadAll(conn)
	test.True(string(bytes) == "hello")

	client, err = mq.NewConnectionWithBroker("test-client", "test", broker)
	if err != nil {
		test.FailNow(err.Error())
		return
	}
	go func() {
		client.Write([]byte(strings.Repeat("hello", 1026)))
		time.Sleep(500 * time.Millisecond)
	}()
	c, err := server.Accept()
	if err != nil {
		test.FailNow(err.Error())
		return
	}

	go func() {
		time.Sleep(500 * time.Millisecond)
		c.Close()
	}()
	raw, _ := ioutil.ReadAll(c)
	test.True(len(raw) > 4096)
	s := string(raw)
	data := strings.ReplaceAll(s, "hello", "")
	spew.Dump(data)
	test.True("" == data)
}
