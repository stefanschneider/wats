package wats

import (
	"fmt"
	"io"
	"net"
	"strings"
	"sync"

	"github.com/cloudfoundry-incubator/cf-test-helpers/cf"
	"github.com/cloudfoundry-incubator/cf-test-helpers/generator"
	"github.com/cloudfoundry-incubator/cf-test-helpers/helpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("Logging", func() {
	var appName string

	FDescribe("Syslog drains", func() {
		var drainListener *syslogDrainListener
		var serviceName string

		BeforeEach(func() {
			drainListener = &syslogDrainListener{}
			drainListener.StartListener()
			go drainListener.AcceptConnections()

			syslogDrainAddress := drainListener.Address()
			fmt.Printf("################ addr: %s\n", syslogDrainAddress)

			testThatDrainIsReachable(syslogDrainAddress, drainListener)

			appName = generator.RandomName()
			pushAndStartNora(appName)

			syslogDrainURL := "syslog://" + syslogDrainAddress
			serviceName = "service-" + generator.RandomName()

			Eventually(cf.Cf("cups", serviceName, "-l", syslogDrainURL), DEFAULT_TIMEOUT).Should(Exit(0), "Failed to create syslog drain service")
			Eventually(cf.Cf("bind-service", appName, serviceName), DEFAULT_TIMEOUT).Should(Exit(0), "Failed to bind service")
		})

		AfterEach(func() {
			Eventually(cf.Cf("delete", appName, "-f"), DEFAULT_TIMEOUT).Should(Exit(0), "Failed to delete app")
			if serviceName != "" {
				Eventually(cf.Cf("delete-service", serviceName, "-f"), DEFAULT_TIMEOUT).Should(Exit(0), "Failed to delete service")
			}

			Eventually(cf.Cf("delete-orphaned-routes", "-f"), CF_PUSH_TIMEOUT).Should(Exit(0), "Failed to delete orphaned routes")

			drainListener.Stop()
		})

		FIt("forwards app messages to registered syslog drains", func() {
			randomMessage := "random-message-" + generator.RandomName()

			Eventually(func() bool {
				helpers.CurlApp(appName, "/print/"+randomMessage)
				return drainListener.DidReceive(randomMessage)
			}, 90, 1).Should(BeTrue(), "Never received "+randomMessage+" on syslog drain listener")
		})
	})
})

type syslogDrainListener struct {
	sync.Mutex
	listener         net.Listener
	receivedMessages string
}

func (s *syslogDrainListener) StartListener() {
	var err error
	s.listener, err = net.Listen("tcp", ":0")
	Expect(err).ToNot(HaveOccurred())
}

func (s *syslogDrainListener) Address() string {
	return s.listener.Addr().String()
}

func (s *syslogDrainListener) AcceptConnections() {
	defer GinkgoRecover()

	for {
		conn, err := s.listener.Accept()
		if err != nil {
			return
		}
		go s.handleConnection(conn)
	}
}

func (s *syslogDrainListener) Stop() {
	s.listener.Close()
}

func (s *syslogDrainListener) DidReceive(message string) bool {
	s.Lock()
	defer s.Unlock()

	return strings.Contains(s.receivedMessages, message)
}

func (s *syslogDrainListener) handleConnection(conn net.Conn) {
	defer GinkgoRecover()
	buffer := make([]byte, 65536)
	for {
		n, err := conn.Read(buffer)

		if err == io.EOF {
			return
		}
		Expect(err).ToNot(HaveOccurred())

		s.Lock()
		s.receivedMessages += string(buffer[0:n])
		s.Unlock()
	}
}

func testThatDrainIsReachable(syslogDrainAddress string, drainListener *syslogDrainListener) {
	conn, err := net.Dial("tcp", syslogDrainAddress)
	Expect(err).ToNot(HaveOccurred())
	defer conn.Close()

	randomMessage := "random-message-" + generator.RandomName()
	_, err = conn.Write([]byte(randomMessage))
	Expect(err).ToNot(HaveOccurred())

	Eventually(func() bool {
		return drainListener.DidReceive(randomMessage)
	}).Should(BeTrue())
}
