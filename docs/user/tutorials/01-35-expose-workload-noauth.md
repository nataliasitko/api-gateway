# Expose a Workload with noAuth

This tutorial shows how to expose an unsecured instance of the HTTPBin Service and call its endpoints using the `noAuth` access strategy.

## Context

The `noAuth` access strategy allows public access to your workload without any authentication or authorization checks. This is useful for:
- Development and testing environments
- Public APIs that don't require authentication
- Services that implement their own authentication logic

> [!WARNING]
> Exposing a workload without authentication is a potential security vulnerability. In production environments, always secure your workloads with proper authentication such as [JWT](./01-40-expose-workload-jwt.md).

## Prerequisites

<!-- tabs:start -->
#### **SAP BTP, Kyma Runtime**

You have Istio and API Gateway modules in your cluster. See [Adding and Deleting a Kyma Module](https://help.sap.com/docs/btp/sap-business-technology-platform/enable-and-disable-kyma-module?locale=en-US&version=Cloud).

#### **k3d**

You have Istio and API Gateway modules in your cluster. See [Quick Install](https://kyma-project.io/02-get-started/01-quick-install.html).

<!-- tabs:end -->

## Steps

>[!NOTE]
> To expose a workload using APIRule in version `v2`, the workload must be part of the Istio service mesh. See [Enable Istio Sidecar Proxy Injection](https://kyma-project.io/external-content/istio/docs/user/tutorials/01-40-enable-sidecar-injection.html#enable-istio-sidecar-proxy-injection).

To expose a workload without authentication, create an APIRule with `noAuth: true` configured for each path you want to expose publicly.

<!-- tabs:start -->
#### **Kyma Dashboard**

In a namespace of your choice, go to **Discovery and Network > API Rules** and choose **Create**. Provide all the required configuration details:

- **Service**: Name and port of your Kubernetes Service
- **Gateway**: Typically `kyma-system/kyma-gateway`
- **Host**: The domain where your workload will be accessible
- **Rules**: Configure with `noAuth` access strategy and specify allowed methods and paths

#### **kubectl**

Replace the placeholders and apply the following configuration. Adjust the rules section as needed.

```bash
cat <<EOF | kubectl apply -f -
apiVersion: gateway.kyma-project.io/v2
kind: APIRule
metadata:
  name: ${APIRULE_NAME}
  namespace: ${NAMESPACE}
spec:
  hosts:
    - ${SUBDOMAIN}.${DOMAIN}
  service:
    name: ${SERVICE_NAME}
    namespace: ${SERVICE_NAMESPACE}
    port: ${SERVICE_PORT}
  gateway: ${GATEWAY_NAMESPACE}/${GATEWAY_NAME}
  rules:
    - path: /post
      methods: ["POST"]
      noAuth: true
    - path: /{**}
      methods: ["GET"]
      noAuth: true
EOF
```

**Placeholders:**

| Placeholder | Description |
|-------------|-------------|
| `${APIRULE_NAME}` | Name for your APIRule resource (for example, `httpbin-noauth`) |
| `${NAMESPACE}` | Namespace where the APIRule will be created |
| `${SUBDOMAIN}.${DOMAIN}` | Full domain where your workload will be accessible |
| `${SERVICE_NAME}` | Name of the Kubernetes Service to expose |
| `${SERVICE_NAMESPACE}` | Namespace where your Service is deployed |
| `${SERVICE_PORT}` | Port on which your Service listens |
| `${GATEWAY_NAMESPACE}/${GATEWAY_NAME}` | Gateway to use (typically `kyma-system/kyma-gateway`) |

<!-- tabs:end -->

## Example

Follow this example to create an APIRule that exposes a sample HTTPBin Deployment.

<!-- tabs:start -->
#### **Kyma Dashboard**

1. Go to **Namespaces** and create a namespace with enabled Istio sidecar proxy injection.

2. Select **+ Upload YAML**, paste the following configuration, and upload it:
    
    ```yaml
    apiVersion: v1
    kind: ServiceAccount
    metadata:
      name: httpbin
    ---
    apiVersion: v1
    kind: Service
    metadata:
      name: httpbin
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
    ```

3. Go to **Discovery and Network > API Rules** and select **Create**.

4. Provide the following details:
   - **Name**: `httpbin-noauth`
   - **Service Name**: `httpbin`
   - **Service Port**: `8000`
   - **Gateway**: `kyma-system/kyma-gateway`

5. Configure the **Host** field based on your environment:

   <!-- tabs:start -->
   #### **SAP BTP, Kyma Runtime**
   
   Add the host `httpbin.${PARENT_DOMAIN}`.
   
   To find your default parent domain, go to the **Kyma Environment** section of your subaccount overview, and copy the part of the **API Server URL** after `https://api.`. For example, if your API Server URL is `https://api.c123abc.kyma.ondemand.com`, use `httpbin.c123abc.kyma.ondemand.com` as the host.
   
   #### **k3d**
   
   Add the host `httpbin.local.kyma.dev`.
   
   This tutorial uses the wildcard public domain `*.local.kyma.dev`, which is registered in public DNS and points to localhost `127.0.0.1`.
   
   <!-- tabs:end -->

6. Add the first rule:
    - **Path**: `/post`
    - **Handler**: `No Auth`
    - **Methods**: `POST`

7. Add the second rule:
    - **Path**: `/{**}`
    - **Handler**: `No Auth`
    - **Methods**: `GET`

8. Choose **Create**.

#### **kubectl**

<!-- tabs:start -->
#### **SAP BTP, Kyma Runtime**

1. Create a namespace with enabled Istio sidecar proxy injection:

    ```bash
    kubectl create ns test
    kubectl label namespace test istio-injection=enabled --overwrite
    ```

2. Get the default domain of your Kyma cluster:

    ```bash
    GATEWAY_DOMAIN=$(kubectl get gateway -n kyma-system kyma-gateway -o jsonpath='{.spec.servers[0].hosts[0]}')
    WORKLOAD_DOMAIN=httpbin.${GATEWAY_DOMAIN#*.}
    GATEWAY=kyma-system/kyma-gateway
    NAMESPACE=test
    
    echo "Workload domain: ${WORKLOAD_DOMAIN}"
    echo "Gateway: ${GATEWAY}"
    ```

3. Deploy a sample instance of the HTTPBin Service:

    ```bash
    cat <<EOF | kubectl -n $NAMESPACE apply -f -
    apiVersion: v1
    kind: ServiceAccount
    metadata:
      name: httpbin
    ---
    apiVersion: v1
    kind: Service
    metadata:
      name: httpbin
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

4. Verify that the HTTPBin pod is running:

    ```bash
    kubectl get pods -l app=httpbin -n $NAMESPACE
    ```

    You should see output similar to:

    ```shell
    NAME                       READY   STATUS    RESTARTS   AGE
    httpbin-5d9d9c9f4b-qj8zw   2/2     Running   0          30s
    ```

5. Expose the workload with an APIRule:

    ```bash
    cat <<EOF | kubectl apply -f -
    apiVersion: gateway.kyma-project.io/v2
    kind: APIRule
    metadata:
      name: httpbin-noauth
      namespace: $NAMESPACE
    spec:
      hosts:
        - ${WORKLOAD_DOMAIN}
      service:
        name: httpbin
        namespace: $NAMESPACE
        port: 8000
      gateway: ${GATEWAY}
      rules:
        - path: /post
          methods: ["POST"]
          noAuth: true
        - path: /{**}
          methods: ["GET"]
          noAuth: true
    EOF
    ```

6. Check if the APIRule is ready:

    ```bash
    kubectl get apirule httpbin-noauth -n $NAMESPACE
    ```

    Wait until the `STATUS` column shows `OK`.

#### **k3d**

1. Create a namespace and export its value as an environment variable:

    ```bash
    export NAMESPACE=api-gateway-tutorial
    kubectl create ns $NAMESPACE
    kubectl label namespace $NAMESPACE istio-injection=enabled --overwrite
    ```

2. Deploy a sample instance of the HTTPBin Service:

    ```bash
    cat <<EOF | kubectl -n $NAMESPACE apply -f -
    apiVersion: v1
    kind: ServiceAccount
    metadata:
      name: httpbin
    ---
    apiVersion: v1
    kind: Service
    metadata:
      name: httpbin
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

3. Verify that the HTTPBin pod is running:

    ```bash
    kubectl get pods -l app=httpbin -n $NAMESPACE
    ```

    You should see output similar to:

    ```shell
    NAME                       READY   STATUS    RESTARTS   AGE
    httpbin-5d9d9c9f4b-qj8zw   2/2     Running   0          30s
    ```

4. Expose the workload with an APIRule:

    ```bash
    cat <<EOF | kubectl apply -f -
    apiVersion: gateway.kyma-project.io/v2
    kind: APIRule
    metadata:
      name: httpbin-noauth
      namespace: $NAMESPACE
    spec:
      hosts:
        - httpbin.local.kyma.dev
      service:
        name: httpbin
        namespace: $NAMESPACE
        port: 8000
      gateway: kyma-system/kyma-gateway
      rules:
        - path: /post
          methods: ["POST"]
          noAuth: true
        - path: /{**}
          methods: ["GET"]
          noAuth: true
    EOF
    ```

5. Check if the APIRule is ready:

    ```bash
    kubectl get apirule httpbin-noauth -n $NAMESPACE
    ```

    Wait until the `STATUS` column shows `OK`.

<!-- tabs:end -->
<!-- tabs:end -->

## Verify the Exposure

Test that your workload is publicly accessible.

<!-- tabs:start -->
#### **SAP BTP, Kyma Runtime**

1. Send a `GET` request to the exposed workload:

    ```bash
    curl -ik -X GET https://${WORKLOAD_DOMAIN}/ip
    ```
  
    If successful, you'll see a `200 OK` response with your IP address.

2. Send a `POST` request to test the POST endpoint:

    ```bash
    curl -ik -X POST https://${WORKLOAD_DOMAIN}/post -d "test data"
    ```
  
    If successful, you'll see a `200 OK` response with the data you sent.

#### **k3d**

1. Send a `GET` request to the exposed workload:

    ```bash
    curl -ik -X GET https://httpbin.local.kyma.dev:30443/ip
    ```
  
    If successful, you'll see a `200 OK` response with your IP address.

2. Send a `POST` request to test the POST endpoint:

    ```bash
    curl -ik -X POST https://httpbin.local.kyma.dev:30443/post -d "test data"
    ```
  
    If successful, you'll see a `200 OK` response with the data you sent.

<!-- tabs:end -->

> [!TIP]
> Your workload is now publicly accessible without any authentication. To secure it, see:
> - [Secure a Workload with JWT](./01-40-expose-workload-jwt.md)
> - [Secure a Workload with OAuth2/OIDC](./01-50-expose-and-secure-a-workload-oauth2.md)
> - [Configure mTLS Authentication](./01-10-mtls-authentication/README.md)
