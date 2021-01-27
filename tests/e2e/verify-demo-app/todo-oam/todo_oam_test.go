package oam

import (
	"fmt"
	"net/http"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "gitlab-odx.oracledx.com/verrazzano/verrazzano-acceptance-test-suite/util"
)

var (
	env                 VerrazzanoEnvironment
	testConfig          VerrazzanoTestConfig
	expectedPodsTodoOam = []string{"tododomain-adminserver"}
	waitTimeout         = 10 * time.Minute
	pollingInterval     = 30 * time.Second
	elastic             *Elastic
)

const (
	ISO8601Layout        = "2006-01-02T15:04:05.999999999-07:00"
	testNamespace        = "todo"
	todoHostHeader       = "todo.example.com"
	testServiceNamespace = "istio-system"
	testServiceName      = "istio-ingressgateway"
)

var _ = BeforeSuite(func() {
	testConfig = GetTestConfig()
	env = NewVerrazzanoEnvironmentFromConfig(testConfig)
	elastic = env.GetElastic("system")
})

var _ = Describe("Verify Todo OAM App.", func() {
	Describe("Verify 'tododomain-adminserver' pod is running.", func() {
		It("and waiting for expected pods must be running", func() {
			Eventually(podsRunningInVerrazzanoApplication, waitTimeout, pollingInterval).Should(BeTrue())
		})
	})

	// These assertions have been commented out because they are failing when we run against a Kind cluster.
	// There are some abstractions currently being worked on that will obtain the Ingress and LB based on cluster type
	// and thee assertions should be changed to use those abstractions.
	//Describe("Verify Todo app is working.", func() {
	//	It("Verify Todo load balancer service exists and ingress IP is assigned.", func() {
	//		Eventually(func() bool {
	//			service, err := env.GetCluster1().ClientSet().CoreV1().Services(testServiceNamespace).Get(testServiceName, metav1.GetOptions{})
	//			return err == nil &&
	//				len(service.Status.LoadBalancer.Ingress) == 1 &&
	//				len(service.Status.LoadBalancer.Ingress[0].IP) > 0
	//		}, 5*time.Minute, 5*time.Second).Should(BeTrue())
	//	})
	//
	//	It("Access /todo App Url.", func() {
	//		Eventually(func() bool {
	//			service, _ := env.GetCluster1().ClientSet().CoreV1().Services(testServiceNamespace).Get(testServiceName, metav1.GetOptions{})
	//			Expect(len(service.Status.LoadBalancer.Ingress)).To(Equal(1))
	//			host := service.Status.LoadBalancer.Ingress[0].IP
	//			url := fmt.Sprintf("http://%s:80/todo", host)
	//			return appEndpointAccessible(url)
	//		}, 5*time.Minute, 5*time.Second).Should(BeTrue())
	//	})
	//})

	Describe("Verify Todo app logging is working.", func() {
		// GIVEN a WLS application with logging enabled via a logging scope
		// WHEN the elastic search index is retrieved
		// THEN verify that it is found
		It("Verify tododomain elasticsearch index exists", func() {
			Eventually(func() bool {
				return logIndexFound("tododomain")
			}, waitTimeout, pollingInterval).Should(BeTrue(), "Expected to find log index tododomain")
		})

		// GIVEN a WLS application with logging enabled via a logging scope
		// WHEN the log records are retrieved from the elastic search index tododomain
		// THEN verify that at least one recent log record is found
		It("Verify recent tododomain elasticsearch log record exists", func() {
			Eventually(func() bool {
				return logRecordFound("tododomain", time.Now().Add(-24*time.Hour),
					Field{"domainUID", "tododomain"},
					Field{"serverName", "tododomain-adminserver"})
			}, waitTimeout, pollingInterval).Should(BeTrue(), "Expected to find a recent log record")
		})
	})

})

func podsRunningInVerrazzanoApplication() bool {
	return env.GetCluster1().Namespace(testNamespace).
		PodsRunning(expectedPodsTodoOam)
}

func appEndpointAccessible(url string) bool {
	status, webpage := GetWebPage(url, todoHostHeader)
	return Expect(status).To(Equal(http.StatusOK), fmt.Sprintf("GET %v returns status %v expected 200.", url, status)) &&
		Expect(len(webpage)).To(Not(Equal(0)), fmt.Sprintf("Return from Todo OAM App is empty string %v.", webpage))
}

func logIndexFound(indexName string) bool {
	for name, _ := range elastic.GetIndices() {
		if name == indexName {
			return true
		}
	}
	fmt.Fprintf(GinkgoWriter, "Expected to find log index %s\n", indexName)
	return false
}

func logRecordFound(indexName string, after time.Time, fields ...Field) bool {
	searchResult := elastic.Search(indexName, fields...)
	hits := jq(searchResult, "hits", "hits")
	if hits == nil {
		fmt.Fprintf(GinkgoWriter, "Expected to find hits in log record query results\n")
		return false
	}
	if len(hits.([]interface{})) == 0 {
		fmt.Fprintf(GinkgoWriter, "Expected log record query results to contain at least one hit\n")
		return false
	}
	for _, hit := range hits.([]interface{}) {
		timestamp := jq(hit, "_source", "@timestamp")
		t, err := time.Parse(ISO8601Layout, timestamp.(string))
		if err != nil {
			return false
		}
		if t.After(after) {
			return true
		}
	}
	fmt.Fprintf(GinkgoWriter, "Expected to find recent log record for index %s\n", indexName)
	return false
}

func jq(node interface{}, path ...string) interface{} {
	for _, p := range path {
		node = node.(map[string]interface{})[p]
	}
	return node
}
