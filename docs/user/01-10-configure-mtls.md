# Configure mTLS Authentication for Your Workloads

Configure mutual TLS (mTLS) authentication for your workloads in SAP BTP, Kyma runtime. Learn how to set up mTLS using Gardener-managed or self-signed certificates, configure Gateways and APIRules for mTLS authentication, and verify the mTLS connection. You can use one of the following approaches:

- Use the default domain of your Kyma runtime cluster.
  
    When you create a SAP BTP, Kyma runtime instance, your cluster receives a default wildcard domain that provides the endpoint for the Kubernetes API server. This is the primary access point for all cluster management operations, used by kubectl and other tools.

    Morover, you can use the default domain to set up an Ingress gateway, and exposed your applications under this host. By default, a simple TLS Gateway `kyma-gateway` is configured under the default wildcard domain of your Kyma cluster. To learn what the domain is, you can check the APIServer URL in your subaccount overview, or fetch the domain name from the default simple TLS Gateway:
    
    ```bash
    kubectl get gateway -n kyma-system kyma-gateway -o jsonpath='{.spec.servers[0].hosts}'
    ```
    
    You can request any subdomain of the assigned default domain and use it to create an mTLS Gateway, as long as it is not used by another resource. For example, if your default domain is `*.c12345.kyma.ondemand.com` you can use such subdomains as `example.c12345.kyma.ondemand.com`, `*.example.c12345.kyma.ondemand.com`, and more.

    To learn how to do this, follow [](#set-up-an-external-dns-provider). The continue with [](#use-gardener-managed-certificates)

- Use your custom domain registered in an external DNS provider.
  
    If you want to expose workloads under a custom domain that is not managed by the default provider but by a custom one, you must first register this DNS provider in your Kyma runtime cluster.

- For local testing, use a k3d cluster and the domain local.kyma.dev. See [](#use-self-signed-certificates).

## Prerequisites
For the scenario with Gardener-managed certificates:
- You have an avaliable domain that you can use for setting up an mTLS Gateway and exposing your workload. You can either use the default domain of your Kyma cluster or a custom domain registered in an external DNS provider.

For the test scenario with self-signed certificates:
- A k3d cluster with Istio and API Gateway modules added.
- OpenSSL

## Set Up an External DNS Provider

## Use Gardener-managed Certificates

## Context


## Procedure

1. Prepare a DNS Entry Pointing to the Istio Gateway IP

   1. Create a Secret with you DNS provider credentials.

        In a namespace of your choice, create a Secret containing credentials for your DNS cloud service provider.
        
        The information you provide to the data field differs depending on the DNS provider you're using. The DNS provider must be supported by Gardener. To learn how to configure the Secret for a specific provider, follow External DNS Management Guidelines.

        ```bash
        cat <<EOF | kubectl apply -f -
        apiVersion: v1
        kind: Secret
        metadata:
        name: dns-credentials
        namespace: {NAMESPACE_NAME}
        type: Opaque
        data:
        {YOUR_CREDENTIALS}
        EOF
        ```

    2. Create a DNSProvider resource that references the Secret with your DNS provider's credentials.

       Option | Description
       ---------|----------
       {NAMESPACE_NAME} | The namespace in which you want to create the DNSProvider resource.
       {PROVIDER_TYPE} | The type of the DNS provider you use. For example, aws-route53, azure-dns, google-clouddns, or other supported provider. For the full list, see External DNS Management and the examples in the external-dns-management repository. 
       {DOMAIN_NAME} | The domain managed by the DNS provider. 

        ```bash
        cat <<EOF | kubectl apply -f -
        apiVersion: dns.gardener.cloud/v1alpha1
        kind: DNSProvider
        metadata:
        name: dns-provider
        namespace: {NAMESPACE}
        annotations:
            dns.gardener.cloud/class: garden
        spec:
        type: {PROVIDER_TYPE}
        secretRef:
            name: dns-credentials
        domains:
            include:
            - {DOMAIN_NAME}
        EOF
        ```

    3. Get the external IP address of the istio-ingressgateway Service in the istio-system namespace and export it in an environment variable:
       
       ```bash
       export IP=$(kubectl -n istio-system get service istio-ingressgateway -o jsonpath='{.status.loadBalancer.ingress[0].ip}') # Assuming only one LoadBalancer with external IP // czy to działa
       ```

   2. Create a DNSEntry resource.
    
        ```bash
        cat <<EOF | kubectl apply -f -
        apiVersion: dns.gardener.cloud/v1alpha1
        kind: DNSEntry
        metadata:
        name: dns-entry
        namespace: {NAMESPACE_NAME}
        annotations:
            dns.gardener.cloud/class: garden
        spec:
        dnsName: "*.{DOMAIN_NAME}"
        ttl: 600
        targets:
            - $IP
        EOF
        ```

2. Create server certificates.
    
    You use a Certificate resource to request, configure, and manage certificates from your Kyma cluster. When you create a Certificate resource in your Cluster, Gardener detects it and starts the process of issuing a certificate. One of Gardener's operators creates an ACME order with Let's Encrypt based on the domain names specified. Let's Encrypt is the default certificate issuer in Kyma. Let's Encrypt provides a challenge to prove that you control the specified domains. Once the challenge is completed successfully, Let's Encrypt issues the certificate. Cert-Manager retrieves the issued certificate and stores it in a Kubernetes Secret as specified in the Certificate resource.

    Option | Description
    ---------|----------
    {TLS_SECRET} | The name of the Secret that Gardener creates. It contains your certificate for the domain specified in the Certificate resource.
    {DOMAIN_NAME} | The domain name for which you request the certificate.

    ```bash
    cat <<EOF | kubectl apply -f -
    apiVersion: cert.gardener.cloud/v1alpha1
    kind: Certificate
    metadata:
    name: gardener-domain-cert
    namespace: istio-system
    spec:
    secretName: {TLS_SECRET}
    commonName: {DOMAIN_NAME}
    issuerRef:
        name: garden
    EOF
    ```
    //Root CA: let's encrypt jest commonly trusted
    // public key servera i private key servera w sekrecie {TLS_SECRET}

3. Create a Secret for the mTLS Gateway.
// sekret ma mieć klucz publiczny servera, prywatny servera i root ca klienta

    ```bash
    kubectl create secret generic -n istio-system kyma-mtls-certs --from-file=cacert=cacert.crt
    ```

4. Create an mTLS Gateway.
 
    ```bash
    cat <<EOF | kubectl apply -f -
    apiVersion: networking.istio.io/v1alpha3
    kind: Gateway
    metadata:
    name: kyma-mtls-gateway
    namespace: default
    spec:
    selector:
        app: istio-ingressgateway
        istio: ingressgateway
    servers:
        - port:
            number: 443
            name: mtls
            protocol: HTTPS
        tls:
            mode: MUTUAL
            credentialName: kyma-mtls-certs
        hosts:
            - "*.{DOMAIN_NAME}"
    EOF
    ```

5. Create an APIRule CR.

    ```bash
    cat <<EOF | kubectl apply -f -
    apiVersion: gateway.kyma-project.io/v2
    kind: APIRule
    metadata:
    name: {APIRULE_NAME}
    namespace: {APIRULE_NAMESPACE}
    spec:
    gateway: {GATEWAY_NAMESPACE}/{GATEWAY_NAME}
    hosts:
        - {DOMAIN_NAME}
    rules:
        - methods:
            - GET
        noAuth: true
        path: /*
        timeout: 300
        request:
            headers:
            X-CLIENT-SSL-CN: '%DOWNSTREAM_PEER_SUBJECT%'
            X-CLIENT-SSL-ISSUER: '%DOWNSTREAM_PEER_ISSUER%'
            X-CLIENT-SSL-SAN: '%DOWNSTREAM_PEER_URI_SAN%'
    service:
        name: {SERVICE_NAME}
        port: {SERVICE_PORT}
    EOF
    ```

6. Connect to the workload.
    
    ```bash
    curl --key ${CLIENT_CERT_KEY_FILE} \
        --cert ${CLIENT_CERT_CRT_FILE} \
        --cacert ${CLIENT_ROOT_CA_CRT_FILE} \
        -ik -X GET https://{SUBDOMAIN}.{DOMAIN}/headers
    ```

### Use k3d and Self-Signed Certificates

## Context
When using self-signed certificates for mTLS, you're creating a certification chain that consists of:
-  A root CA certificates for the client and the server that you create (acting as your own Certificate Authority)
- Server and client certificates that are signed by the respective root CA

This means you're establishing trust relationships without using a publicly trusted authority. Therefore, this approach is recommended for use in testing or development environments only. For production deployments, use trusted certificate authorities to ensure proper security and automatic certificate management (for example, Let's Encrypt, as shown in the previous section).

## Prerequisites
- A k3d cluster with Istio and API Gateway modules added. See Quick Start.
- OpenSSL

## Procedure

1. Create the server's root CA.

    ```bash
    export SERVER_ROOT_CA_CN="ML Server Root CA"
    export SERVER_ROOT_CA_ORG="ML Server Org"
    export SERVER_ROOT_CA_KEY_FILE=${SERVER_ROOT_CA_CN}.key
    export SERVER_ROOT_CA_CRT_FILE=${SERVER_ROOT_CA_CN}.crt
    openssl req -x509 -sha256 -nodes -days 365 -newkey rsa:2048 -subj "/O=${SERVER_ROOT_CA_ORG}/CN=${SERVER_ROOT_CA_CN}" -keyout "${SERVER_ROOT_CA_KEY_FILE}" -out "${SERVER_ROOT_CA_CRT_FILE}"
    ```
2. Create the server's certificate.
    
    ```bash
    export SERVER_CERT_CN="httpbin.local.kyma.dev"
    export SERVER_CERT_ORG="ML Server Org"
    export SERVER_CERT_CRT_FILE=${SERVER_CERT_CN}.crt
    export SERVER_CERT_CSR_FILE=${SERVER_CERT_CN}.csr
    export SERVER_CERT_KEY_FILE=${SERVER_CERT_CN}.key
    openssl req -out "${SERVER_CERT_CSR_FILE}" -newkey rsa:2048 -nodes -keyout "${SERVER_CERT_KEY_FILE}" -subj "/CN=${SERVER_CERT_CN}/O=${SERVER_CERT_ORG}"
    ```
3. Sign the server's certificate.
    ```bash
    openssl x509 -req -days 365 -CA "${SERVER_ROOT_CA_CRT_FILE}" -CAkey "${SERVER_ROOT_CA_KEY_FILE}" -set_serial 0 -in "${SERVER_CERT_CSR_FILE}" -out "${SERVER_CERT_CRT_FILE}"
    ```
4. Create client's root CA.
    
    ```bash 
    export CLIENT_ROOT_CA_CN="ML Client Root CA"
    export CLIENT_ROOT_CA_ORG="ML Client Org"
    export CLIENT_ROOT_CA_KEY_FILE=${CLIENT_ROOT_CA_CN}.key
    export CLIENT_ROOT_CA_CRT_FILE=${CLIENT_ROOT_CA_CN}.crt
    openssl req -x509 -sha256 -nodes -days 365 -newkey rsa:2048 -subj "/O=${CLIENT_ROOT_CA_ORG}/CN=${CLIENT_ROOT_CA_CN}" -keyout "${CLIENT_ROOT_CA_KEY_FILE}" -out "${CLIENT_ROOT_CA_CRT_FILE}"
    ```
5. Create the client's certificate.
    
    ```bash
    export CLIENT_CERT_CN="ML Client Curl"
    export CLIENT_CERT_ORG="ML Client Org"
    export CLIENT_CERT_CRT_FILE=${CLIENT_CERT_CN}.crt
    export CLIENT_CERT_CSR_FILE=${CLIENT_CERT_CN}.csr
    export CLIENT_CERT_KEY_FILE=${CLIENT_CERT_CN}.key
    openssl req -out "${CLIENT_CERT_CSR_FILE}" -newkey rsa:2048 -nodes -keyout "${CLIENT_CERT_KEY_FILE}" -subj "/CN=${CLIENT_CERT_CN}/O=${CLIENT_CERT_ORG}"
    ```

6. Sign the client's certificate.
    
    ```bash
    openssl x509 -req -days 365 -CA "${CLIENT_ROOT_CA_CRT_FILE}" -CAkey "${CLIENT_ROOT_CA_KEY_FILE}" -set_serial 0 -in "${CLIENT_CERT_CSR_FILE}" -out "${CLIENT_CERT_CRT_FILE}"
    ```

7. Create a Secret for the mTLS Gateway.
    
    ```bash
    kubectl create secret generic -n istio-system kyma-mtls-certs --from-file=cacert="${CLIENT_ROOT_CA_CRT_FILE}"  --from-file=key="${SERVER_CERT_KEY_FILE}" --from-file=cert="${SERVER_CERT_CRT_FILE}"
    ```
8. Create the mTLS Gateway.
    
    ```bash
    cat <<EOF | kubectl apply -f -
    apiVersion: networking.istio.io/v1alpha3
    kind: Gateway
    metadata:
      name: kyma-mtls-gateway
      namespace: test
    spec:
      selector:
        app: istio-ingressgateway
        istio: ingressgateway
      servers:
        - port:
            number: 443
            name: mtls
            protocol: HTTPS
          tls:
            mode: MUTUAL
            credentialName: kyma-mtls-certs
          hosts:
            - "httpbin.local.kyma.dev"
    EOF
    ```
9. Create a HTTPBin Deployment.

    ```bash
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
    ```

10. Create an APIRule.
    
    ```bash
    cat <<EOF | kubectl apply -f -
    apiVersion: gateway.kyma-project.io/v2
    kind: APIRule
    metadata:
      name: httpbin-mtls
      namespace: test
    spec:
      gateway: test/kyma-mtls-gateway
      hosts:
        - "httpbin.local.kyma.dev"
      rules:
        - methods:
            - GET
          noAuth: true
          path: /*
          timeout: 300
          request:
            headers:
              X-CLIENT-SSL-CN: '%DOWNSTREAM_PEER_SUBJECT%'
              X-CLIENT-SSL-ISSUER: '%DOWNSTREAM_PEER_ISSUER%'
              X-CLIENT-SSL-SAN: '%DOWNSTREAM_PEER_URI_SAN%'
      service:
        name: httpbin
        port: 8000
    EOF
    ```

11. Test the connection.
    
    ```bash
    curl --verbose  \
        --key "${CLIENT_CERT_KEY_FILE}" \
        --cert "${CLIENT_CERT_CRT_FILE}" \
        --cacert "${SERVER_ROOT_CA_CRT_FILE}" \
        "https://httpbin.local.kyma.dev/headers?show_env=true"
    ```