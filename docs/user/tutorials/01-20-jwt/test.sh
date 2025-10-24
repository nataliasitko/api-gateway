# Domain managed by GCP (there must be a Zone in Cloud DNS for it)
PARENT_DOMAIN=goat.build.kyma-project.io
CLUSTER_NAME=tlstest
# Subdomain of the above domain
SUBDOMAIN="${CLUSTER_NAME}.${PARENT_DOMAIN}"

# Domain to be registered, may be a wildcard subdomain of the above subdomain
GATEWAY_DOMAIN="*.${SUBDOMAIN}"

# Domain used by the workload
WORKLOAD_DOMAIN="httpbin.${SUBDOMAIN}"

# Key file in json format downloaded from Service Accounts / Keys in the Google Console UI
DNS_SECRET_JSON_PATH=/Users/I583356/Downloads/sap-se-cx-kyma-goat-80d1b03d4249.json
DNS_SECRET_JSON_B64="$(cat "${DNS_SECRET_JSON_PATH}" | base64)"

# KUBECONFIG=/Users/I583356/Desktop/config-goat\ copy.yaml
CLUSTER_KUBECONFIG=/Users/I583356/Desktop/trial.yaml

export KUBECONFIG="${CLUSTER_KUBECONFIG}"

echo "Delete test namespace if already exists"
kubectl delete ns test

echo "Create test namespace"
kubectl create ns test
kubectl label namespace test istio-injection=enabled --overwrite

echo "Create secret for DNS Provider"
cat <<EOF | kubectl apply -f -
apiVersion: v1
kind: Secret
metadata:
  name: google-dns-provider-credentials
  namespace: test
type: Opaque
data:
  serviceaccount.json: "${DNS_SECRET_JSON_B64}"
EOF

echo "Create DNS Provider"
cat <<EOF | kubectl apply -f -
apiVersion: dns.gardener.cloud/v1alpha1
kind: DNSProvider
metadata:
  name: google-dns-provider
  namespace: test
  annotations:
    dns.gardener.cloud/class: garden
spec:
  type: "google-clouddns"
  secretRef:
    name: "google-dns-provider-credentials"
  domains:
    include:
      - "${PARENT_DOMAIN}"
EOF

echo "Wait until DNS Provider is ready"
kubectl wait dnsprovider -n test google-dns-provider --for=jsonpath='{.status.state}'=Ready --timeout=120s

echo "Get Load Balancer address"
LOAD_BALANCER_ADDRESS=$(kubectl get services --namespace istio-system istio-ingressgateway --output jsonpath='{.status.loadBalancer.ingress[0].ip}')
if [ "$LOAD_BALANCER_ADDRESS" == "" ]; then
  echo "Load Balancer IP address not found, get the host name instead"
  LOAD_BALANCER_ADDRESS=$(kubectl get services --namespace istio-system istio-ingressgateway --output jsonpath='{.status.loadBalancer.ingress[0].hostname}')
fi
if [ "$LOAD_BALANCER_ADDRESS" == "" ]; then
  echo "Can't get Load Balancer address!"
  exit 1
fi
echo "Load Balancer address: ${LOAD_BALANCER_ADDRESS}"

echo "Create DNS entry"
cat <<EOF | kubectl apply -f -
apiVersion: dns.gardener.cloud/v1alpha1
kind: DNSEntry
metadata:
  name: dns-entry
  namespace: test
  annotations:
    dns.gardener.cloud/class: garden
spec:
  dnsName: "${GATEWAY_DOMAIN}"
  ttl: 600
  targets:
    - "${LOAD_BALANCER_ADDRESS}"
EOF

echo "Wait until DNS Entry is ready"
kubectl wait dnsentry -n test dns-entry --for=jsonpath='{.status.state}'=Ready --timeout=600s

echo "Create certificate for the domain ${GATEWAY_DOMAIN}"
GATEWAY_SECRET=my-tls-secret
cat <<EOF | kubectl apply -f -
apiVersion: cert.gardener.cloud/v1alpha1
kind: Certificate
metadata:
  name: domain-certificate
  namespace: "istio-system"
spec:
  secretName: "${GATEWAY_SECRET}"
  commonName: "${GATEWAY_DOMAIN}"
  issuerRef:
    name: garden
EOF

echo "Wait until certificate is ready"
kubectl wait certificate -n istio-system domain-certificate --for=condition=ready --timeout=600s

echo "Secret for TLS Gateway with server key and cert (created by Gardener)"
kubectl get secret -n istio-system "${GATEWAY_SECRET}"

echo "Create TLS Gateway"
cat <<EOF | kubectl apply -f -
apiVersion: networking.istio.io/v1alpha3
kind: Gateway
metadata:
  name: kyma-tls-gateway
  namespace: test
spec:
  selector:
    app: istio-ingressgateway
    istio: ingressgateway
  servers:
    - port:
        number: 443
        name: tls
        protocol: HTTPS
      tls:
        mode: SIMPLE
        credentialName: "${GATEWAY_SECRET}"
      hosts:
        - "${GATEWAY_DOMAIN}"
EOF

echo "Encode client credentials"
export IDENTITY_AUTHENTICATION_INSTANCE="ag2ppojhf.trial-accounts.ondemand.com"
export CLIENT_ID="d5694e36-6a63-4201-8226-d2397356d531"
export CLIENT_SECRET="l?R@[PbsJ?zi3.mxqY3seff/HJWWvFWPsO"
export ENCODED_CREDENTIALS=$(echo -n "$CLIENT_ID:$CLIENT_SECRET" | base64)


echo "Get token_endpoint"
TOKEN_ENDPOINT=$(curl -s https://$IDENTITY_AUTHENTICATION_INSTANCE/.well-known/openid-configuration | jq -r '.token_endpoint')
echo token_endpoint: $TOKEN_ENDPOINT

echo "Get jwks_uri"
JWKS_URI=$(curl -s https://$IDENTITY_AUTHENTICATION_INSTANCE/.well-known/openid-configuration | jq -r '.jwks_uri')
echo jwks_uri: $JWKS_URI

echo "Get issuer"
ISSUER=$(curl -s https://$IDENTITY_AUTHENTICATION_INSTANCE/.well-known/openid-configuration | jq -r '.issuer')
echo issuer: $ISSUER

echo "Get JWT access token"
response=$(curl -s -X POST "$TOKEN_ENDPOINT" \
    -d "grant_type=client_credentials" \
    -d "client_id=$CLIENT_ID" \
    -H "Content-Type: application/x-www-form-urlencoded" \
    -H "Authorization: Basic $ENCODED_CREDENTIALS")

access_token=$(echo $response | jq -r '.access_token')
echo JWT: $access_token

echo "Create Deployment"
cat <<EOF | kubectl apply -f -
apiVersion: v1
kind: ServiceAccount
metadata:
  name: httpbin
  namespace: test
---
apiVersion: v1
kind: Service
metadata:
  name: httpbin
  namespace: test
  labels:
    app: httpbin
    service: httpbin
spec:
  ports:
  - name: http
    port: 8000
    targetPort: 80
  selector:
    app: httpbin
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: httpbin
  namespace: test
spec:
  replicas: 1
  selector:
    matchLabels:
      app: httpbin
      version: v1
  template:
    metadata:
      labels:
        app: httpbin
        version: v1
    spec:
      serviceAccountName: httpbin
      containers:
      - image: docker.io/kennethreitz/httpbin
        imagePullPolicy: IfNotPresent
        name: httpbin
        ports:
        - containerPort: 80
EOF

echo "Wait until deployment is ready"
sleep 10
kubectl wait pods -n test -l app=httpbin --for=condition=ready --timeout=120s

echo "Create APIRule"
cat <<EOF | kubectl apply -f -
apiVersion: gateway.kyma-project.io/v2
kind: APIRule
metadata:
  name: httpbin-tls
  namespace: test
spec:
  gateway: test/kyma-tls-gateway
  hosts:
    - "${WORKLOAD_DOMAIN}"
  rules:
    - jwt:
        authentications:
          - issuer: ${ISSUER}
            jwksUri: ${JWKS_URI}
      methods:
        - GET
      path: /*
  service:
    name: httpbin
    port: 8000
EOF

echo "Wait until APIRule is ready"
kubectl wait -n test apirule httpbin-tls --for=jsonpath='{.status.state}=Ready' --timeout=60s

echo "Call APIRule without JWT"
curl -ik -X GET https://${WORKLOAD_DOMAIN}/headers

echo "Call APIRule using JWT"
curl -ik -X GET https://${WORKLOAD_DOMAIN}/headers --header "Authorization:Bearer $ACCESS_TOKEN"


#HTTP/2 200
#server: istio-envoy
#date: Thu, 23 Oct 2025 13:53:42 GMT
#content-type: application/json
#content-length: 1790
#x-envoy-upstream-service-time: 2
#
#{
#  "headers": {
#    "Accept": "*/*",
#    "Authorization": "Bearer eyJqa3UiOiJodHRwczovL2FnMnBwb2poZi50cmlhbC1hY2NvdW50cy5vbmRlbWFuZC5jb20vb2F1dGgyL2NlcnRzIiwia2lkIjoiUzZYZ0NjOTdVSDJ0Rjl5emZrbmI2WE1XZHQ0IiwiYWxnIjoiUlMyNTYifQ.eyJzdWIiOiJkNTY5NGUzNi02YTYzLTQyMDEtODIyNi1kMjM5NzM1NmQ1MzEiLCJhdWQiOiJkNTY5NGUzNi02YTYzLTQyMDEtODIyNi1kMjM5NzM1NmQ1MzEiLCJhenAiOiJkNTY5NGUzNi02YTYzLTQyMDEtODIyNi1kMjM5NzM1NmQ1MzEiLCJpc3MiOiJodHRwczovL2FnMnBwb2poZi50cmlhbC1hY2NvdW50cy5vbmRlbWFuZC5jb20iLCJhenBhY3IiOiIxIiwiZXhwIjoxNzYxMjMwNzMwLCJpYXQiOjE3NjEyMjcxMzAsImp0aSI6IjYyMzQ2OGEzLTE1ZjEtNDJmZi1iNDBjLTg5ZmQ1NzEyYWQzMyJ9.UT7OIhGa777QnAPkyp2yVSK6ACe3g-aX2TOQT5xbC7nt1vKeGUpfU63dyy8q0ffxzphsFI53pAm02V8dbLGw67_WHz_ujy2odTbAlKvrp4TJWXRQ9NYIce_ETmcQi89S239vu05iN7X67NmUZDF44Zq53ou1aOnh4TXuIiWE8wB2o6s_awTj6dZrenbfaTzMOs0DjN1eGGxXzZRdZOg__rEYlWIFTfKYk-kQx_vjgcjMtX10ha4GFicqLE9GsBi1LcMKTFm2-yOWL-Cl5nuZx6-EdLpIMPKCEBH9eZT1inPMpi6KC69gzSTuXOL_WK7UYi1Pp_zBPjMotJS61WcAEF80zwoti-wS-S1c0rTTi5SgdWc9oFp3iXYFdR2UQaekcoHsL6CcJIVpGKBQFwSsZup0SGNugdpUM6xWzsx9Wi0lAJwqDQnhG1VoBYODNxIbz_fzkp0EDUPDexwR6petWeuiiVJphelYJOdVzCIAJ-P285xW53dyFHWHkMdzC_hoWy6sCkOU6LbnCjYS0n0J05Ubszdbq3SFSGetmJQRkQoVafxXmZ4B7koLdq05Pq-D3BYoIG-P1zE6oPXxbNj3oI64VrR_Wg-ozmFviIcOtxRj06AJ7l3HU8SI5zKjtNOo-tAiXN48qYQhdgKPENPZN-9fZiKmqOOPfS-dZV-wGmk",
#    "Host": "httpbin.tlstest.goat.build.kyma-project.io",
#    "User-Agent": "curl/8.7.1",
#    "X-Envoy-Attempt-Count": "1",
#    "X-Envoy-External-Address": "240.243.202.99",
#    "X-Forwarded-Client-Cert": "By=spiffe://cluster.local/ns/test/sa/httpbin;Hash=899e90d2110eee31d13302ed80252c90be4abd066263c123a7afe5922f14a261;Subject=\"\";URI=spiffe://cluster.local/ns/istio-system/sa/istio-ingressgateway-service-account",
#    "X-Forwarded-Host": "httpbin.tlstest.goat.build.kyma-project.io"
#  }
#}
