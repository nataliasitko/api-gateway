# Quick Start for SAP BTP, Kyma Runtime

This tutorial is dedicated for SAP BTP, Kyma runtime users. Follow the steps to get started with the API Gateway module.

## Prerequisites

- You have access to Kyma dashboard. Alternatively, to use CLI instructions, you must install [kubectl](https://help.sap.com/docs/btp/sap-business-technology-platform-internal/access-kyma-instance-using-kubectl?locale=en-US&state=DRAFT&version=Internal&comment_id=22217515&show_comments=true) and [curl](https://curl.se/).
- You have added the Istio and API Gateway modules to your SAP BTP, Kyma runtime instance. See [Add and Delete ...]().
- You have prepared a domain for exposing a sample workload. Although using a custom domain is highly recommended in production environments, for the purpose of this tutorial, you can use a Kyma domain instead. To learn what is the domain of your Kyma cluster, run ... .

## Context
This Quick Start guide shows how to:
- create a sample HTTPBin workload,
- expose the sample workload to the internet using the APIRule custom resource (CR),
- secure the workload with a JWT token obtained using SAP Cloud Identity Services. 

## Procedure

## Create a Workload

<!-- tabs:start -->
#### **Kyma Dashboard**

1. In Kyma dashboard, go to **Namespaces** and choose **Create**.
1. Use the name `api-gateway-tutorial` and switch the toggle to enable Istio sidecar proxy injection.
2. Choose **Create**.
3. In the created namespace, go to **Workloads > Deployemnts** and choose **Create**.
1. Select the HTTPBin template.
2. Choose **Create**.
3. Go to **Configuration > Service Accounts** and choose **Create**. 
4. Enter `httpbin` as your Service Account's name.
5. Choose **Create**.
6. Go to **Discovery and Network > Services** and choose **Create**. 
7. Provide the following configuration details:
    - **Name**: `httpbin`
    - In the `Labels` section, add the following labels:
      - **service**: `httpbin`
      - **app**:`httpbin`
    - In the `Selectors` section, add the following selector:
      - **app**: `httpbin`
    - In the `Ports` section, select **Add**. Then, use these values:
      - **Name**: `http`
      - **Protocol**: `TCP`
      - **Port**: `8000`
      - **Target Port**: `80`
8. Choose **Create**.

#### **kubectl**

1. Create a namespace and export its value as an environment variable. Run:

    ```bash
    export NAMESPACE=api-gateway-tutorial
    kubectl create ns $NAMESPACE
    kubectl label namespace $NAMESPACE istio-injection=enabled --overwrite
    ```

2. Deploy a sample instance of the HTTPBin Service.

    ```shell
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

    To verify if an instance of the HTTPBin Service is successfully created, run:

    ```shell
    kubectl get pods -l app=httpbin -n $NAMESPACE
    ```

    If successful, you get a result similar to this one:

    ```shell
    NAME                 READY    STATUS     RESTARTS    AGE
    httpbin-{SUFFIX}     2/2      Running    0           96s
    ```

<!-- tabs:end -->

## Expose a Workload

<!-- tabs:start -->
#### **Kyma Dashboard**

1. In  `api-gateway-tutorial` namespace, go to **Discovery and Network > API Rules**.
2. Choose **Create**.
3. Provide the following configuration details.
  - **Name**: `httpbin`
  - In the `Service` section, select:
    - **Service Name**: `httpbin`
    - **Port**: `8000`
  - In the `Service` section, add:
    - **Namespace**: `api-gateway-tutorial`
    - **Name**: `kyma-gateway`
    - In the **Host** filed, add `httpbin.{YOUR_DOMAIN}`. Replace the placeholder with the name of your domain.
  - In the `Rules` section, add two Rules. Use the following configuration for the first one:
    - **Path**: `/.*`
    - **Handler**: `no_auth`
    - **Methods**: `GET`
  - Use the following configuration for the second Rule:
    - **Path**: `/post`
    - **Handler**: `no_auth`
    - **Methods**: `POST`
4.  Choose **Create**.

#### **kubectl**

1. Export the name of your domain as an environment variable:

  ```bash
  export DOMAIN_TO_EXPOSE_WORKLOADS={DOMAIN_NAME}
  ```

2. To expose the HTTPBin Service, create the follwing APIRule CR. Run:

```bash
cat <<EOF | kubectl apply -f -
apiVersion: gateway.kyma-project.io/v1beta1
kind: APIRule
metadata:
  name: httpbin
  namespace: api-gateway-tutorial
spec:
  host: httpbin.{DOMAIN_TO_EXPOSE_WORKLOADS}
  service:
    name: httpbin
    namespace: api-gateway-tutorial
    port: 8000
  gateway: kyma-gateway
  rules:
    - path: /.*
      methods: ["GET"]
      accessStrategies:
        - handler: no_auth
    - path: /post
      methods: ["POST"]
      accessStrategies:
        - handler: no_auth
EOF
```

<!-- tabs:end -->

## Access a Workload

To access the HTTPBin Service, use [curl](https://curl.se).

- Send a `GET` request to the HTTPBin Service.

  ```bash
  curl -ik -X GET https://httpbin.local.kyma.dev:30443/ip
  ```
  If successful, the call returns the `200 OK` response code.

- Send a `POST` request to the HTTPBin Service.

  ```bash
  curl -ik -X POST https://httpbin.local.kyma.dev:30443/post -d "test data"
  ```
  If successful, the call returns the `200 OK` response code.

<!-- tabs:end -->