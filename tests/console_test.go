package tests_test

import (
	"flag"
	"net/url"

	"strings"
	"time"

	"github.com/gorilla/websocket"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"k8s.io/client-go/pkg/api"

	"kubevirt.io/kubevirt/pkg/api/v1"
	"kubevirt.io/kubevirt/pkg/kubecli"
	"kubevirt.io/kubevirt/tests"
)

var _ = Describe("Vmlifecycle", func() {

	flag.Parse()

	restClient, err := kubecli.GetRESTClient()
	tests.PanicOnError(err)
	var vm *v1.VM
	var dial func(vm string, console string) *websocket.Conn

	BeforeEach(func() {
		tests.MustCleanup()

		vm = tests.NewRandomVMWithSerialConsole()

		dial = func(vm string, console string) *websocket.Conn {
			wsUrl, err := url.Parse(flag.Lookup("master").Value.String())
			Expect(err).ToNot(HaveOccurred())
			wsUrl.Scheme = "ws"
			wsUrl.Path = "/apis/kubevirt.io/v1alpha1/namespaces/default/vms/" + vm + "/console"
			wsUrl.RawQuery = "console=" + console
			c, _, err := websocket.DefaultDialer.Dial(wsUrl.String(), nil)
			Expect(err).ToNot(HaveOccurred())
			return c
		}
	})

	Context("New VM with a serial console given", func() {

		It("should be allowed to connect to the console", func(done Done) {
			Expect(restClient.Post().Resource("vms").Namespace(api.NamespaceDefault).Body(vm).Do().Error()).To(Succeed())
			tests.WaitForSuccessfulVMStart(vm)
			ws := dial(vm.ObjectMeta.GetName(), "serial0")
			defer ws.Close()
			close(done)
		}, 60)

		It("should be returned that we are running cirros", func(done Done) {
			Expect(restClient.Post().Resource("vms").Namespace(api.NamespaceDefault).Body(vm).Do().Error()).To(Succeed())
			tests.WaitForSuccessfulVMStart(vm)
			ws := dial(vm.ObjectMeta.GetName(), "serial0")
			defer ws.Close()
			// Check for the typical cloud init error messages
			// TODO, use a reader instead and use ReadLine from bufio
			next := ""
			Eventually(func() string {
				for {
					if index := strings.Index(next, "\n"); index != -1 {
						line := next[0:index]
						next = next[index+1:]
						return line
					}
					_, data, err := ws.ReadMessage()
					Expect(err).ToNot(HaveOccurred())
					next = next + string(data)
				}
			}, 60*time.Second).Should(ContainSubstring("checking http://169.254.169.254/2009-04-04/instance-id"))
			close(done)
		}, 90)

		AfterEach(func() {
			tests.MustCleanup()
		})
	})
})
