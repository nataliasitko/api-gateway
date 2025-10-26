# TLS Authentication

This tutorial shows how to set up a TLS Gateway in simple mode.

## Prerequisites

* [Set up your custom domain](./01-10-setup-custom-domain-for-workload.md).

## Context
- You have a SAP BTP, Kyma runtime instance with Istio and API Gateway modules added. The Istio and API Gateway modules are added to your Kyma cluster by default.
- For setting up the mTLS Gateway, you must prepare the domain name available in the public DNS zone. You can use one of the following approaches:

  - Use your custom domain.

    For a custom domain you must own the DNS zone and supply credentials for a provider supported by Gardener so the ACME DNS challenge can be completed. For this, you must first register this DNS provider in your Kyma runtime cluster and create a DNS entry resource.

  - Use the default domain of your Kyma cluster.

    When you create a SAP BTP, Kyma runtime instance, your cluster receives a default wildcard domain that provides the endpoint for the Kubernetes API server. This is the primary access point for all cluster management operations, used by kubectl and other tools.

    By default, the default Ingress Gateway `kyma-gateway` is configured under this domain. To learn what the domain is, you can check the APIServer URL in your subaccount overview, or get the domain name from the default simple TLS Gateway: 
    ```bash
    kubectl get gateway -n kyma-system kyma-gateway -o jsonpath='{.spec.servers[0].hosts}'
    ```

    You can request any subdomain of the assigned default domain and use it to create a TLS or mTLS Gateway, as long as it is not used by another resource. For example, if your default domain is `*.c12345.kyma.ondemand.com` you can use such subdomains as `example.c12345.kyma.ondemand.com`, `*.example.c12345.kyma.ondemand.com`, and more. If you use the Kyma runtime default domain, Gardener’s issuer can issue certificates for subdomains of that domain without additional DNS delegation.

## Steps

<!-- tabs:start -->
### **Custom Domain**

1. Create a namespace with enabled Istio sidecar proxy injection.

    ```bash
    kubectl create ns test
    kubectl label namespace test istio-injection=enabled --overwrite
    ```
2. Export the following domain names as enviroment variables. Replace `my-own-domain.example.com` with the name of your domain.

    ```bash
    PARENT_DOMAIN="my-own-domain.example.com"
    SUBDOMAIN="mtls.${PARENT_DOMAIN}"
    GATEWAY_DOMAIN="*.${SUBDOMAIN}"
    WORKLOAD_DOMAIN="httpbin.${SUBDOMAIN}"
    ```

    Placeholder | Example domain name | Description
    ---------|----------|---------
    **PARENT_DOMAIN** | `my-own-domain.example.com` | The domain name available in the public DNS zone.
    **SUBDOMAIN** | `mtls.my-own-domain.example.com` | A subdomain created under the parent domain, specifically for the mTLS Gateway.
    **GATEWAY_DOMAIN** | `*.mtls.my-own-domain.example.com` | A wildcard domain covering all possible subdomains under the mTLS subdomain. When configuring the Gateway, this allows you to expose workloads on multiple hosts (for example, `httpbin.mtls.my-own-domain.example.com`, `test.httpbin.mtls.my-own-domain.example.com`) without creating separate Gateway rules for each one.
    **WORKLOAD_DOMAIN** | `httpbin.mtls.my-own-domain.example.com` | The specific domain assigned to your workload.

3. Create a Secret containing credentials for your DNS cloud service provider.

    The information you provide to the data field differs depending on the DNS provider you're using. The DNS provider must be supported by Gardener. To learn how to configure the Secret for a specific provider, follow [External DNS Management Guidelines](https://github.com/gardener/cert-management?tab=readme-ov-file#using-commonname-and-optional-dnsnames).
    See an example Secret for AWS Route 53 DNS provider. **AWS_ACCESS_KEY_ID** and **AWS_SECRET_ACCESS_KEY** are base-64 encoded credentials.
    ```bash
    apiVersion: v1
    kind: Secret
    metadata:
      name: aws-credentials
      namespace: test
    type: Opaque
    data:
      AWS_ACCESS_KEY_ID: ...
      AWS_SECRET_ACCESS_KEY: ...
      # Optionally, specify the region
      #AWS_REGION: {YOUR_SECRET_ACCESS_KEY
      # Optionally, specify the token
      #AWS_SESSION_TOKEN: ...
    EOF
    ```
4. Create a DNSProvider resource that references the Secret with your DNS provider's credentials.

   See an example Secret for AWS Route 53 DNS provider:

    ```bash
    cat <<EOF | kubectl apply -f -
    apiVersion: dns.gardener.cloud/v1alpha1
    kind: DNSProvider
    metadata:
      name: aws
      namespace: default
    spec:
      type: aws-route53
      ecretRef:
        name: aws-credentials
      domains:
        include:
        - "${PARENT_DOMAIN}"
    EOF
    ```
5. Get the external access point of the `istio-ingressgateway` Service. The external access point is either stored in the ingress Gateway's **ip** field (for example, on GCP) or in the ingress Gateway's **hostname** field (for example, on AWS).
    ```bash
    LOAD_BALANCER_ADDRESS=$(kubectl get services --namespace istio-system istio-ingressgateway --output jsonpath='{.status.loadBalancer.ingress[0].ip}')
    if [[ -z $LOAD_BALANCER_ADDRESS ]]; then
        LOAD_BALANCER_ADDRESS=$(kubectl get services --namespace istio-system istio-ingressgateway --output jsonpath='{.status.loadBalancer.ingress[0].hostname}')
    fi
    ```
6. Create a DNSEntry resource.
    
    ```bash
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
    ```
7. Create the server's certificate.
    
    You use a Certificate resource to request and manage Let's Encrypt certificates from your Kyma cluster. When you create a Certificate, Gardener detects it and starts the process of issuing a certificate. One of Gardener's operators detects it and creates an ACME order with Let's Encrypt based on the domain names specified. Let's Encrypt is the default certificate issuer in Kyma. Let's Encrypt provides a challenge to prove that you own the specified domains. Once the challenge is completed successfully, Let's Encrypt issues the certificate. The issued certificate is stored it in a Kubernetes Secret, which name is specified in the Certificate's **secretName** field.
    ```bash
    export GATEWAY_SECRET=kyma-mtls
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
    ```
    To verify that the Scret with Gateway certificates is created, run:
   
    ```bash
    kubectl get secret -n istio-system "${GATEWAY_SECRET}"
    ```

8.  Create a TLS Gateway.
 
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
            credentialName: "${GATEWAY_SECRET}"
          hosts:
            - "${GATEWAY_DOMAIN}"
    EOF
    ```
    
### **Default Domain**
1. Create a namespace with enabled Istio sidecar proxy injection.
   
    ```bash
    kubectl create ns test
    kubectl label namespace test istio-injection=enabled --overwrite
    ```
2. Export the following domain names as enviroment variables. Replace `my-own-domain.kyma.ondemand.com` with the name of your domain.
    ```bash
    PARENT_DOMAIN="my-own-domain.kyma.ondemand.com"
    SUBDOMAIN="mtls.${PARENT_DOMAIN}"
    GATEWAY_DOMAIN="*.${SUBDOMAIN}"
    WORKLOAD_DOMAIN="httpbin.${SUBDOMAIN}"
    ```
    Placeholder | Example domain name | Description
    ---------|----------|---------
    **PARENT_DOMAIN** | `my-default-domain.kyma.ondemand.com` | The default domain of your Kyma cluster.
    **SUBDOMAIN** | `mtls.my-default-domain.kyma.ondemand.com` | A subdomain created under the parent domain, specifically for the mTLS Gateway. Choosing a subdomain is required if you use the default domain of your Kyma cluster, as the parent domain name is already assigned to the TLS Gateway `kyma-gateway` installed in your cluster by default.
    **GATEWAY_DOMAIN** | `*.mtls.my-default-domain.kyma.ondemand.com` | A wildcard domain covering all possible subdomains under the mTLS subdomain. When configuring the Gateway, this allows you to expose workloads on multiple hosts (for example, `httpbin.mtls.my-default-domain.kyma.ondemand.com`, `test.httpbin.mtls.my-default-domain.kyma.ondemand.com`) without creating separate Gateway rules for each one.
    **WORKLOAD_DOMAIN** | `httpbin.mtls.my-default-domain.kyma.ondemand.com` | The specific domain assigned to your sample workload (HTTPBin service) in this tutorial.
3. Create the server's certificate.
    
    You use a Certificate resource to request and manage Let's Encrypt certificates from your Kyma cluster. When you create a Certificate, Gardener detects it and starts the process of issuing a certificate. One of Gardener's operators detects it and creates an ACME order with Let's Encrypt based on the domain names specified. Let's Encrypt is the default certificate issuer in Kyma. Let's Encrypt provides a challenge to prove that you own the specified domains. Once the challenge is completed successfully, Let's Encrypt issues the certificate. The issued certificate is stored it in a Kubernetes Secret, which name is specified in the Certificate's **secretName** field.
    ```bash
    export GATEWAY_SECRET=kyma-mtls
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
    ```
    To verify that the Scret with Gateway certificates is created, run:
   
    ```bash
    kubectl get secret -n istio-system "${GATEWAY_SECRET}"
    ```

4.  Create a TLS Gateway.
 
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
            credentialName: "${GATEWAY_SECRET}"
          hosts:
            - "${GATEWAY_DOMAIN}"
    EOF
    ```
5.  Create a sample HTTPBin Deployment.
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
6.  To expose the sample HTTPBin Deployment, create an APIRule custom resource. 
    
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
        - "${WORKLOAD_DOMAIN}"
      rules:
        - methods:
            - GET
          noAuth: true
          path: /*
      service:
        name: httpbin
        port: 8000
    EOF
    ```
7.  Test the TLS connection.
    
    1. Run the following curl command:
    
        ```bash
        curl --fail --verbose \
          "https://${WORKLOAD_DOMAIN}/headers?show_env==true"
        ```
        
        If successful, you get code `200` in response.
    
    2. To thest the connection using your browser, open `https://{WORKLOAD_DOMAIN}`.
<!-- tabs:end -->

<!-- tabs:start -->
#### **Kyma Dashboard**

1. Go to **Istio > Gateways** and select **Create**.
2. Provide the following configuration details:
    - **Name**: `example-gateway`
    - Add a server with the following values:
      - **Port Number**: `443`
      - **Name**: `https`
      - **Protocol**: `HTTPS`
      - **TLS Mode**: `SIMPLE`
      - **Credential Name** is the name of the Secret that contains the credentials.
    - Use `{SUBDOMAIN}.{CUSTOM_DOMAIN}` as a host.

3. Select **Create**.

#### **kubectl**

1. Export the following values as environment variables:

    ```bash
    export DOMAIN_TO_EXPOSE_WORKLOADS={DOMAIN_NAME}
    export NAMESPACE={YOUR_NAMESPACE}
    export GATEWAY=$NAMESPACE/example-gateway
    ```

2. To create a TLS Gateway in simple mode, run:

    ```bash
cat <<EOF | kubectl apply -f -
---
apiVersion: networking.istio.io/v1alpha3
kind: Gateway
metadata:
  name: example-gateway
  namespace: $NAMESPACE
spec:
  selector:
    istio: ingressgateway
  servers:
    - port:
        number: 443
        name: https
        protocol: HTTPS
      tls:
        mode: SIMPLE
        credentialName: $TLS_SECRET
      hosts:
        - "*.$DOMAIN_TO_EXPOSE_WORKLOADS"
EOF
    ```

<!-- tabs:end -->