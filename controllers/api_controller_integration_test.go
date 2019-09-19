package controllers_test

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"encoding/json"

	gatewayv1alpha1 "github.com/kyma-incubator/api-gateway/api/v1alpha1"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	rulev1alpha1 "github.com/ory/oathkeeper-maester/api/v1alpha1"
	"k8s.io/apimachinery/pkg/api/errors"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	networkingv1alpha3 "knative.dev/pkg/apis/istio/v1alpha3"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

const (
	timeout = time.Second * 5

	kind                        = "APIRule"
	testGatewayURL              = "kyma-gateway.kyma-system.svc.cluster.local"
	testOathkeeperSvcURL        = "oathkeeper.kyma-system.svc.cluster.local"
	testOathkeeperPort   uint32 = 1234
	testNamespace               = "padu-system"
	testNameBase                = "test"
	testIDLength                = 5
)

var _ = Describe("APIRule Controller", func() {
	const testServiceName = "httpbin"
	const testServiceHost = "httpbin.kyma.local"
	const testServicePort uint32 = 443
	const testPath = "/.*"
	var testIssuer = "https://oauth2.example.com/"
	var testMethods = []string{"GET", "PUT"}
	var testScopes = []string{"foo", "bar"}
	var testMutators = []*rulev1alpha1.Mutator{
		{
			Handler: &rulev1alpha1.Handler{
				Name: "noop",
			},
		},
		{
			Handler: &rulev1alpha1.Handler{
				Name: "idtoken",
			},
		},
	}

	Context("when creating an APIRule for exposing service", func() {

		It("Should report validation errors in CR status", func() {
			configJSON := fmt.Sprintf(`{
							"required_scope": [%s]
						}`, toCSVList(testScopes))

			nonEmptyConfig := &rulev1alpha1.Handler{
				Name: "noop",
				Config: &runtime.RawExtension{
					Raw: []byte(configJSON),
				},
			}

			testName := generateTestName(testNameBase, testIDLength)
			rule := testRule(testPath, testMethods, testMutators, nonEmptyConfig)
			instance := testInstance(testName, testNamespace, testServiceName, testServiceHost, testServicePort, []gatewayv1alpha1.Rule{rule})
			instance.Spec.Rules = append(instance.Spec.Rules, instance.Spec.Rules[0]) //Duplicate entry
			instance.Spec.Rules = append(instance.Spec.Rules, instance.Spec.Rules[0]) //Duplicate entry

			err := c.Create(context.TODO(), instance)
			if apierrors.IsInvalid(err) {
				Fail(fmt.Sprintf("failed to create object, got an invalid object error: %v", err))
				return
			}
			Expect(err).NotTo(HaveOccurred())
			defer c.Delete(context.TODO(), instance)

			expectedRequest := reconcile.Request{NamespacedName: types.NamespacedName{Name: testName, Namespace: testNamespace}}

			Eventually(requests, timeout).Should(Receive(Equal(expectedRequest)))

			//Verify APIRule
			created := gatewayv1alpha1.APIRule{}
			err = c.Get(context.TODO(), client.ObjectKey{Name: testName, Namespace: testNamespace}, &created)
			Expect(err).NotTo(HaveOccurred())
			Expect(created.Status.APIRuleStatus.Code).To(Equal(gatewayv1alpha1.StatusError))
			Expect(created.Status.APIRuleStatus.Description).To(ContainSubstring("Multiple validation errors:"))
			Expect(created.Status.APIRuleStatus.Description).To(ContainSubstring("Attribute \".spec.rules\": multiple rules defined for the same path"))
			Expect(created.Status.APIRuleStatus.Description).To(ContainSubstring("Attribute \".spec.rules[0].accessStrategies[0].config\": strategy: noop does not support configuration"))
			Expect(created.Status.APIRuleStatus.Description).To(ContainSubstring("Attribute \".spec.rules[1].accessStrategies[0].config\": strategy: noop does not support configuration"))
			Expect(created.Status.APIRuleStatus.Description).To(ContainSubstring("1 more error(s)..."))

			//Verify VirtualService is not created
			expectedVSName := testName + "-" + testServiceName
			expectedVSNamespace := testNamespace
			vs := networkingv1alpha3.VirtualService{}
			err = c.Get(context.TODO(), client.ObjectKey{Name: expectedVSName, Namespace: expectedVSNamespace}, &vs)
			Expect(errors.IsNotFound(err)).To(BeTrue())
		})

		Context("on all the paths,", func() {
			Context("secured with Oauth2 introspection,", func() {
				Context("in a happy-path scenario", func() {
					It("should create a VirtualService and an AccessRule", func() {
						configJSON := fmt.Sprintf(`{
							"required_scope": [%s]
						}`, toCSVList(testScopes))

						oauthConfig := &rulev1alpha1.Handler{
							Name: "oauth2_introspection",
							Config: &runtime.RawExtension{
								Raw: []byte(configJSON),
							},
						}

						testName := generateTestName(testNameBase, testIDLength)
						rule := testRule(testPath, testMethods, testMutators, oauthConfig)
						instance := testInstance(testName, testNamespace, testServiceName, testServiceHost, testServicePort, []gatewayv1alpha1.Rule{rule})

						err := c.Create(context.TODO(), instance)
						if apierrors.IsInvalid(err) {
							Fail(fmt.Sprintf("failed to create object, got an invalid object error: %v", err))
							return
						}
						Expect(err).NotTo(HaveOccurred())
						defer c.Delete(context.TODO(), instance)

						expectedRequest := reconcile.Request{NamespacedName: types.NamespacedName{Name: testName, Namespace: testNamespace}}

						Eventually(requests, timeout).Should(Receive(Equal(expectedRequest)))

						//Verify VirtualService
						expectedVSName := testName + "-" + testServiceName
						expectedVSNamespace := testNamespace
						vs := networkingv1alpha3.VirtualService{}
						err = c.Get(context.TODO(), client.ObjectKey{Name: expectedVSName, Namespace: expectedVSNamespace}, &vs)
						Expect(err).NotTo(HaveOccurred())

						//Meta
						verifyOwnerReference(vs.ObjectMeta, testName, gatewayv1alpha1.GroupVersion.String(), kind)
						//Spec.Hosts
						Expect(vs.Spec.Hosts).To(HaveLen(1))
						Expect(vs.Spec.Hosts[0]).To(Equal(testServiceHost))
						//Spec.Gateways
						Expect(vs.Spec.Gateways).To(HaveLen(1))
						Expect(vs.Spec.Gateways[0]).To(Equal(testGatewayURL))
						//Spec.HTTP
						Expect(vs.Spec.HTTP).To(HaveLen(1))
						////// HTTP.Match[]
						Expect(vs.Spec.HTTP[0].Match).To(HaveLen(1))
						/////////// Match[].URI
						Expect(vs.Spec.HTTP[0].Match[0].URI).NotTo(BeNil())
						Expect(vs.Spec.HTTP[0].Match[0].URI.Exact).To(BeEmpty())
						Expect(vs.Spec.HTTP[0].Match[0].URI.Prefix).To(BeEmpty())
						Expect(vs.Spec.HTTP[0].Match[0].URI.Suffix).To(BeEmpty())
						Expect(vs.Spec.HTTP[0].Match[0].URI.Regex).To(Equal(testPath))
						Expect(vs.Spec.HTTP[0].Match[0].Scheme).To(BeNil())
						Expect(vs.Spec.HTTP[0].Match[0].Method).To(BeNil())
						Expect(vs.Spec.HTTP[0].Match[0].Authority).To(BeNil())
						Expect(vs.Spec.HTTP[0].Match[0].Headers).To(BeNil())
						Expect(vs.Spec.HTTP[0].Match[0].Port).To(BeZero())
						Expect(vs.Spec.HTTP[0].Match[0].SourceLabels).To(BeNil())
						Expect(vs.Spec.HTTP[0].Match[0].Gateways).To(BeNil())
						////// HTTP.Route[]
						Expect(vs.Spec.HTTP[0].Route).To(HaveLen(1))
						Expect(vs.Spec.HTTP[0].Route[0].Destination.Host).To(Equal(testOathkeeperSvcURL))
						Expect(vs.Spec.HTTP[0].Route[0].Destination.Subset).To(Equal(""))
						Expect(vs.Spec.HTTP[0].Route[0].Destination.Port.Name).To(Equal(""))
						Expect(vs.Spec.HTTP[0].Route[0].Destination.Port.Number).To(Equal(testOathkeeperPort))
						Expect(vs.Spec.HTTP[0].Route[0].Weight).To(BeZero())
						Expect(vs.Spec.HTTP[0].Route[0].Headers).To(BeNil())
						//Others
						Expect(vs.Spec.HTTP[0].Rewrite).To(BeNil())
						Expect(vs.Spec.HTTP[0].WebsocketUpgrade).To(BeFalse())
						Expect(vs.Spec.HTTP[0].Timeout).To(BeEmpty())
						Expect(vs.Spec.HTTP[0].Retries).To(BeNil())
						Expect(vs.Spec.HTTP[0].Fault).To(BeNil())
						Expect(vs.Spec.HTTP[0].Mirror).To(BeNil())
						Expect(vs.Spec.HTTP[0].DeprecatedAppendHeaders).To(BeNil())
						Expect(vs.Spec.HTTP[0].Headers).To(BeNil())
						Expect(vs.Spec.HTTP[0].RemoveResponseHeaders).To(BeNil())
						Expect(vs.Spec.HTTP[0].CorsPolicy).To(BeNil())
						//Spec.TCP
						Expect(vs.Spec.TCP).To(BeNil())
						//Spec.TLS
						Expect(vs.Spec.TLS).To(BeNil())

						//Verify Rule
						expectedRuleName := testName + "-" + testServiceName + "-0"
						expectedRuleNamespace := testNamespace
						rl := rulev1alpha1.Rule{}
						err = c.Get(context.TODO(), client.ObjectKey{Name: expectedRuleName, Namespace: expectedRuleNamespace}, &rl)
						Expect(err).NotTo(HaveOccurred())

						//Meta
						verifyOwnerReference(rl.ObjectMeta, testName, gatewayv1alpha1.GroupVersion.String(), kind)

						//Spec.Upstream
						Expect(rl.Spec.Upstream).NotTo(BeNil())
						Expect(rl.Spec.Upstream.URL).To(Equal(fmt.Sprintf("http://%s.%s.svc.cluster.local:%d", testServiceName, testNamespace, testServicePort)))
						Expect(rl.Spec.Upstream.StripPath).To(BeNil())
						Expect(rl.Spec.Upstream.PreserveHost).To(BeNil())
						//Spec.Match
						Expect(rl.Spec.Match).NotTo(BeNil())
						Expect(rl.Spec.Match.URL).To(Equal(fmt.Sprintf("<http|https>://%s<%s>", testServiceHost, testPath)))
						Expect(rl.Spec.Match.Methods).To(Equal(testMethods))
						//Spec.Authenticators
						Expect(rl.Spec.Authenticators).To(HaveLen(1))
						Expect(rl.Spec.Authenticators[0].Handler).NotTo(BeNil())
						Expect(rl.Spec.Authenticators[0].Handler.Name).To(Equal("oauth2_introspection"))
						Expect(rl.Spec.Authenticators[0].Handler.Config).NotTo(BeNil())
						//Authenticators[0].Handler.Config validation
						handlerConfig := map[string]interface{}{}
						err = json.Unmarshal(rl.Spec.Authenticators[0].Config.Raw, &handlerConfig)
						Expect(err).NotTo(HaveOccurred())
						Expect(handlerConfig).To(HaveLen(1))
						Expect(asStringSlice(handlerConfig["required_scope"])).To(BeEquivalentTo(testScopes))
						//Spec.Authorizer
						Expect(rl.Spec.Authorizer).NotTo(BeNil())
						Expect(rl.Spec.Authorizer.Handler).NotTo(BeNil())
						Expect(rl.Spec.Authorizer.Handler.Name).To(Equal("allow"))
						Expect(rl.Spec.Authorizer.Handler.Config).To(BeNil())

						//Spec.Mutators
						Expect(rl.Spec.Mutators).NotTo(BeNil())
						Expect(len(rl.Spec.Mutators)).To(Equal(len(testMutators)))
						Expect(rl.Spec.Mutators[0].Handler.Name).To(Equal(testMutators[0].Name))
						Expect(rl.Spec.Mutators[1].Handler.Name).To(Equal(testMutators[1].Name))
					})
				})
			})
			Context("secured with JWT token authentication,", func() {
				Context("in a happy-path scenario", func() {
					It("should create a VirtualService and an AccessRules", func() {
						configJSON := fmt.Sprintf(`
							{
								"trusted_issuers": ["%s"],
								"jwks": [],
								"required_scope": [%s]
						}`, testIssuer, toCSVList(testScopes))
						jwtConfig := &rulev1alpha1.Handler{
							Name: "jwt",
							Config: &runtime.RawExtension{
								Raw: []byte(configJSON),
							},
						}
						testName := generateTestName(testNameBase, testIDLength)
						rule1 := testRule("/img", []string{"GET"}, testMutators, jwtConfig)
						rule2 := testRule("/headers", []string{"GET"}, testMutators, jwtConfig)
						instance := testInstance(testName, testNamespace, testServiceName, testServiceHost, testServicePort, []gatewayv1alpha1.Rule{rule1, rule2})

						err := c.Create(context.TODO(), instance)
						if apierrors.IsInvalid(err) {
							Fail(fmt.Sprintf("failed to create object, got an invalid object error: %v", err))
							return
						}
						Expect(err).NotTo(HaveOccurred())
						defer c.Delete(context.TODO(), instance)

						expectedRequest := reconcile.Request{NamespacedName: types.NamespacedName{Name: testName, Namespace: testNamespace}}

						Eventually(requests, timeout).Should(Receive(Equal(expectedRequest)))
						//Verify VirtualService
						expectedVSName := testName + "-" + testServiceName
						expectedVSNamespace := testNamespace
						vs := networkingv1alpha3.VirtualService{}
						err = c.Get(context.TODO(), client.ObjectKey{Name: expectedVSName, Namespace: expectedVSNamespace}, &vs)
						Expect(err).NotTo(HaveOccurred())

						//Meta
						verifyOwnerReference(vs.ObjectMeta, testName, gatewayv1alpha1.GroupVersion.String(), kind)
						//Spec.Hosts
						Expect(vs.Spec.Hosts).To(HaveLen(1))
						Expect(vs.Spec.Hosts[0]).To(Equal(testServiceHost))
						//Spec.Gateways
						Expect(vs.Spec.Gateways).To(HaveLen(1))
						Expect(vs.Spec.Gateways[0]).To(Equal(testGatewayURL))
						//Spec.HTTP
						Expect(vs.Spec.HTTP).To(HaveLen(1))
						////// HTTP.Match[]
						Expect(vs.Spec.HTTP[0].Match).To(HaveLen(1))
						/////////// Match[].URI
						Expect(vs.Spec.HTTP[0].Match[0].URI).NotTo(BeNil())
						Expect(vs.Spec.HTTP[0].Match[0].URI.Exact).To(BeEmpty())
						Expect(vs.Spec.HTTP[0].Match[0].URI.Prefix).To(BeEmpty())
						Expect(vs.Spec.HTTP[0].Match[0].URI.Suffix).To(BeEmpty())
						Expect(vs.Spec.HTTP[0].Match[0].URI.Regex).To(Equal("/.*"))
						Expect(vs.Spec.HTTP[0].Match[0].Scheme).To(BeNil())
						Expect(vs.Spec.HTTP[0].Match[0].Method).To(BeNil())
						Expect(vs.Spec.HTTP[0].Match[0].Authority).To(BeNil())
						Expect(vs.Spec.HTTP[0].Match[0].Headers).To(BeNil())
						Expect(vs.Spec.HTTP[0].Match[0].Port).To(BeZero())
						Expect(vs.Spec.HTTP[0].Match[0].SourceLabels).To(BeNil())
						Expect(vs.Spec.HTTP[0].Match[0].Gateways).To(BeNil())
						////// HTTP.Route[]
						Expect(vs.Spec.HTTP[0].Route).To(HaveLen(1))
						Expect(vs.Spec.HTTP[0].Route[0].Destination.Host).To(Equal(testOathkeeperSvcURL))
						Expect(vs.Spec.HTTP[0].Route[0].Destination.Subset).To(Equal(""))
						Expect(vs.Spec.HTTP[0].Route[0].Destination.Port.Name).To(Equal(""))
						Expect(vs.Spec.HTTP[0].Route[0].Destination.Port.Number).To(Equal(testOathkeeperPort))
						Expect(vs.Spec.HTTP[0].Route[0].Weight).To(BeZero())
						Expect(vs.Spec.HTTP[0].Route[0].Headers).To(BeNil())
						//Others
						Expect(vs.Spec.HTTP[0].Rewrite).To(BeNil())
						Expect(vs.Spec.HTTP[0].WebsocketUpgrade).To(BeFalse())
						Expect(vs.Spec.HTTP[0].Timeout).To(BeEmpty())
						Expect(vs.Spec.HTTP[0].Retries).To(BeNil())
						Expect(vs.Spec.HTTP[0].Fault).To(BeNil())
						Expect(vs.Spec.HTTP[0].Mirror).To(BeNil())
						Expect(vs.Spec.HTTP[0].DeprecatedAppendHeaders).To(BeNil())
						Expect(vs.Spec.HTTP[0].Headers).To(BeNil())
						Expect(vs.Spec.HTTP[0].RemoveResponseHeaders).To(BeNil())
						Expect(vs.Spec.HTTP[0].CorsPolicy).To(BeNil())
						//Spec.TCP
						Expect(vs.Spec.TCP).To(BeNil())
						//Spec.TLS
						Expect(vs.Spec.TLS).To(BeNil())

						//Verify Rule1
						expectedRuleName := testName + "-" + testServiceName + "-0"
						expectedRuleNamespace := testNamespace
						rl := rulev1alpha1.Rule{}
						err = c.Get(context.TODO(), client.ObjectKey{Name: expectedRuleName, Namespace: expectedRuleNamespace}, &rl)
						Expect(err).NotTo(HaveOccurred())

						//Meta
						verifyOwnerReference(rl.ObjectMeta, testName, gatewayv1alpha1.GroupVersion.String(), kind)

						//Spec.Upstream
						Expect(rl.Spec.Upstream).NotTo(BeNil())
						Expect(rl.Spec.Upstream.URL).To(Equal(fmt.Sprintf("http://%s.%s.svc.cluster.local:%d", testServiceName, testNamespace, testServicePort)))
						Expect(rl.Spec.Upstream.StripPath).To(BeNil())
						Expect(rl.Spec.Upstream.PreserveHost).To(BeNil())
						//Spec.Match
						Expect(rl.Spec.Match).NotTo(BeNil())
						Expect(rl.Spec.Match.URL).To(Equal(fmt.Sprintf("<http|https>://%s<%s>", testServiceHost, "/img")))
						Expect(rl.Spec.Match.Methods).To(Equal([]string{"GET"}))
						//Spec.Authenticators
						Expect(rl.Spec.Authenticators).To(HaveLen(1))
						Expect(rl.Spec.Authenticators[0].Handler).NotTo(BeNil())
						Expect(rl.Spec.Authenticators[0].Handler.Name).To(Equal("jwt"))
						Expect(rl.Spec.Authenticators[0].Handler.Config).NotTo(BeNil())
						//Authenticators[0].Handler.Config validation
						handlerConfig := map[string]interface{}{}

						err = json.Unmarshal(rl.Spec.Authenticators[0].Config.Raw, &handlerConfig)
						Expect(err).NotTo(HaveOccurred())
						Expect(handlerConfig).To(HaveLen(3))
						Expect(asStringSlice(handlerConfig["required_scope"])).To(BeEquivalentTo(testScopes))
						Expect(asStringSlice(handlerConfig["trusted_issuers"])).To(BeEquivalentTo([]string{testIssuer}))
						//Spec.Authorizer
						Expect(rl.Spec.Authorizer).NotTo(BeNil())
						Expect(rl.Spec.Authorizer.Handler).NotTo(BeNil())
						Expect(rl.Spec.Authorizer.Handler.Name).To(Equal("allow"))
						Expect(rl.Spec.Authorizer.Handler.Config).To(BeNil())

						//Spec.Mutators
						Expect(rl.Spec.Mutators).NotTo(BeNil())
						Expect(len(rl.Spec.Mutators)).To(Equal(len(testMutators)))
						Expect(rl.Spec.Mutators[0].Handler.Name).To(Equal(testMutators[0].Name))
						Expect(rl.Spec.Mutators[1].Handler.Name).To(Equal(testMutators[1].Name))

						//Verify Rule2
						expectedRuleName2 := testName + "-" + testServiceName + "-1"
						rl2 := rulev1alpha1.Rule{}
						err = c.Get(context.TODO(), client.ObjectKey{Name: expectedRuleName2, Namespace: expectedRuleNamespace}, &rl2)
						Expect(err).NotTo(HaveOccurred())

						//Meta
						verifyOwnerReference(rl2.ObjectMeta, testName, gatewayv1alpha1.GroupVersion.String(), "APIRule")

						//Spec.Upstream
						Expect(rl2.Spec.Upstream).NotTo(BeNil())
						Expect(rl2.Spec.Upstream.URL).To(Equal(fmt.Sprintf("http://%s.%s.svc.cluster.local:%d", testServiceName, testNamespace, testServicePort)))
						Expect(rl2.Spec.Upstream.StripPath).To(BeNil())
						Expect(rl2.Spec.Upstream.PreserveHost).To(BeNil())
						//Spec.Match
						Expect(rl2.Spec.Match).NotTo(BeNil())
						Expect(rl2.Spec.Match.URL).To(Equal(fmt.Sprintf("<http|https>://%s<%s>", testServiceHost, "/headers")))
						Expect(rl2.Spec.Match.Methods).To(Equal([]string{"GET"}))
						//Spec.Authenticators
						Expect(rl2.Spec.Authenticators).To(HaveLen(1))
						Expect(rl2.Spec.Authenticators[0].Handler).NotTo(BeNil())
						Expect(rl2.Spec.Authenticators[0].Handler.Name).To(Equal("jwt"))
						Expect(rl2.Spec.Authenticators[0].Handler.Config).NotTo(BeNil())
						//Authenticators[0].Handler.Config validation
						handlerConfig = map[string]interface{}{}

						err = json.Unmarshal(rl2.Spec.Authenticators[0].Config.Raw, &handlerConfig)
						Expect(err).NotTo(HaveOccurred())
						Expect(handlerConfig).To(HaveLen(3))
						Expect(asStringSlice(handlerConfig["required_scope"])).To(BeEquivalentTo(testScopes))
						Expect(asStringSlice(handlerConfig["trusted_issuers"])).To(BeEquivalentTo([]string{testIssuer}))
						//Spec.Authorizer
						Expect(rl2.Spec.Authorizer).NotTo(BeNil())
						Expect(rl2.Spec.Authorizer.Handler).NotTo(BeNil())
						Expect(rl2.Spec.Authorizer.Handler.Name).To(Equal("allow"))
						Expect(rl2.Spec.Authorizer.Handler.Config).To(BeNil())

						//Spec.Mutators
						Expect(rl2.Spec.Mutators).NotTo(BeNil())
						Expect(len(rl2.Spec.Mutators)).To(Equal(len(testMutators)))
						Expect(rl2.Spec.Mutators[0].Handler.Name).To(Equal(testMutators[0].Name))
						Expect(rl2.Spec.Mutators[1].Handler.Name).To(Equal(testMutators[1].Name))
					})
				})
			})
		})
	})
})

func add(mgr manager.Manager, r reconcile.Reconciler) error {
	c, err := controller.New("api-gateway-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	err = c.Watch(&source.Kind{Type: &gatewayv1alpha1.APIRule{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	return nil
}

func toCSVList(input []string) string {
	if len(input) == 0 {
		return ""
	}

	res := `"` + input[0] + `"`

	for i := 1; i < len(input); i++ {
		res = res + "," + `"` + input[i] + `"`
	}

	return res
}

func testRule(path string, methods []string, mutators []*rulev1alpha1.Mutator, config *rulev1alpha1.Handler) gatewayv1alpha1.Rule {
	return gatewayv1alpha1.Rule{
		Path:     path,
		Methods:  methods,
		Mutators: mutators,
		AccessStrategies: []*rulev1alpha1.Authenticator{
			{
				Handler: config,
			},
		},
	}
}

func testInstance(name, namespace, serviceName, serviceHost string, servicePort uint32, rules []gatewayv1alpha1.Rule) *gatewayv1alpha1.APIRule {
	var gateway = testGatewayURL

	return &gatewayv1alpha1.APIRule{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: gatewayv1alpha1.APIRuleSpec{
			Gateway: &gateway,
			Service: &gatewayv1alpha1.Service{
				Host: &serviceHost,
				Name: &serviceName,
				Port: &servicePort,
			},
			Rules: rules,
		},
	}
}

func verifyOwnerReference(m metav1.ObjectMeta, name, version, kind string) {
	Expect(m.OwnerReferences).To(HaveLen(1))
	Expect(m.OwnerReferences[0].APIVersion).To(Equal(version))
	Expect(m.OwnerReferences[0].Kind).To(Equal(kind))
	Expect(m.OwnerReferences[0].Name).To(Equal(name))
	Expect(m.OwnerReferences[0].UID).NotTo(BeEmpty())
	Expect(*m.OwnerReferences[0].Controller).To(BeTrue())
}

//Converts a []interface{} to a string slice. Panics if given object is of other type.
func asStringSlice(in interface{}) []string {

	inSlice := in.([]interface{})

	if inSlice == nil {
		return nil
	}

	res := []string{}

	for _, v := range inSlice {
		res = append(res, v.(string))
	}

	return res
}

func generateTestName(name string, length int) string {

	rand.Seed(time.Now().UnixNano())

	letterRunes := []rune("abcdefghijklmnopqrstuvwxyz")

	b := make([]rune, length)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return name + "-" + string(b)
}