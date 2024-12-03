# Quick Start: Expose and Secure a Sample Workload

This tutorial is dedicated for SAP BTP, Kyma runtime users. Follow the steps to get started with the API Gateway module.

## Prerequisites

- You have access to Kyma dashboard. Alternatively, to use CLI instructions, you must install [kubectl](https://help.sap.com/docs/btp/sap-business-technology-platform-internal/access-kyma-instance-using-kubectl?locale=en-US&state=DRAFT&version=Internal&comment_id=22217515&show_comments=true) and [curl](https://curl.se/).
- You have added the Istio and API Gateway modules to your SAP BTP, Kyma runtime instance. See [Add and Delete ...]().
- You have prepared a domain for exposing a sample workload.

For the purpose of this tutorial, you can use a Kyma domain instead of your custom domain. If you use a k3d cluster, your kyma domain is `local.kyma.dev`. If you use a Gardener cluster, you can check the domain by running ... .

## Context
This Quick Start guide shows how to:
- create a sample HTTPBin workload,
- expose the sample workload to the internet using the APIRule custom resource (CR),
- secure the workload using SAP Cloud Identity Services.

## Procedure

<!-- tabs:start -->
#### **Kyma Dashboard**

1. Create a namespace with enabled Istio sidecar proxy injection.
2. Go to **Workloads > Deployments**.
3. Choose **Create**. 
4. Select the HTTPBin template.
5. Choose **Create**.

#### **kubectl**

1. Create a namespace and export its value as an environment variable. Run:

    ```bash
    export NAMESPACE={NAMESPACE_NAME}
    kubectl create ns $NAMESPACE
    kubectl label namespace $NAMESPACE istio-injection=enabled --overwrite
    ```

2. Choose a name for your HTTPBin Service instance and export it as an environment variable.

    ```bash
    export SERVICE_NAME={SERVICE_NAME}
    ```

3. Deploy a sample instance of the HTTPBin Service.

    ```shell
    cat <<EOF | kubectl -n $NAMESPACE apply -f -
    apiVersion: v1
    kind: ServiceAccount
    metadata:
      name: $SERVICE_NAME
    ---
    apiVersion: v1
    kind: Service
    metadata:
      name: $SERVICE_NAME
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
      name: $SERVICE_NAME
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
          serviceAccountName: $SERVICE_NAME
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

    You should get a result similar to this one:

    ```shell
    NAME                        READY    STATUS     RESTARTS    AGE
    {SERVICE_NAME}-{SUFFIX}     2/2      Running    0           96s
    ```

<!-- tabs:end -->
