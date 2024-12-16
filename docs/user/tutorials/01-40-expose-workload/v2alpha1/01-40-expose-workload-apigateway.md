# Expose a Workload

This tutorial shows how to expose an unsecured instance of the HTTPBin Service and call its endpoints.

> [!WARNING]
>  Exposing a workload to the outside world is a potential security vulnerability, so tread carefully. In a production environment, always secure the workload you expose with [JWT](../../01-50-expose-and-secure-a-workload/v2alpha1/01-52-expose-and-secure-workload-jwt.md).

## Prerequisites

- You have deployed a Service.
- You have set up your custom domain and a TLS Gateway. See [Set Up Your Custom Domain](../../01-10-setup-custom-domain-for-workload.md) and Set Up a TLS Gateway. Alternatively, you can use the default domain of your Kyma cluster and the default Gateway `kyma-system/kyma-gateway`. To check the name of your cluster's Kyma domain, run: ...

## Steps

### Expose Your Workload

  <!-- tabs:start -->
  #### **Kyma Dashboard**

  1. Go to **Discovery and Network > API Rules**.
  2. Choose **Create**.
  3. Provide the following configuration details.
    - Add the APIRule's name in the **Name** field.
    - In the `Service` section, add the Service's name and port.
    - To fill in the `Gateway` section, use these values:
      - **Namespace** is the name of the namespace in which you deployed an instance of the HTTPBin Service. If you use a Kyma domain, select the `kyma-system` namespace.
      - **Name** is the Gateway's name. If you use a Kyma domain, select `kyma-gateway`.
      - In the **Host** field, enter `httpbin.{DOMAIN_TO_EXPORT_WORKLOADS}`. Replace the placeholder with the name of your domain.
    - In the `Rules` section, add two Rules. Use the following configuration for the first one:
      - **Path**: `/.*`
      - **Handler**: `no_auth`
      - **Methods**: `GET`
    - Use the following configuration for the second Rule:
      - **Path**: `/post`
      - **Handler**: `no_auth`
      - **Methods**: `POST`

  4. To create the APIRule, select **Create**.
  
  #### **kubectl**

  1. Export the name of your Service and its namespace:

      ```bash
      export SERVICE={SERVICE_NAME}
      export NAMESPACE={NAMESPACE_NAME}
      ```

  2. Export the name of your domain to expose workloads and the Gateway:

      ```bash
      export DOMAIN_TO_EXPOSE_WORKLOADS={DOMAIN_NAME}
      export GATEWAY={GATEWAY_NAMESPACE}/{GATEWAY_NAME}
      ```

  3. To expose your Service, create an APIRule CR. You can adjust the following configuration according to your needs. For more information, see APIRule Custom Resource.

      ```bash
      cat <<EOF | kubectl apply -f -
      apiVersion: gateway.kyma-project.io/v2alpha1
      kind: APIRule
      metadata:
        name: httpbin
        namespace: $NAMESPACE
      spec:
        hosts:
          - httpbin.$DOMAIN_TO_EXPOSE_WORKLOADS
        service:
          name: $SERVICE_NAME
          namespace: $NAMESPACE
          port: 8000
        gateway: $GATEWAY
        rules:
          - path: /*
            methods: ["GET"]
            noAuth: true
          - path: /post
            methods: ["POST"]
            noAuth: true
      EOF
      ```

  <!-- tabs:end -->

### Access Your Workload

To access your HTTPBin Service, [curl](https://curl.se).

- Send a `GET` request to the HTTPBin Service.

  ```bash
  curl -ik -X GET https://httpbin.$DOMAIN_TO_EXPOSE_WORKLOADS/ip
  ```
  If successful, the call returns the `200 OK` response code.

- Send a `POST` request to the HTTPBin Service.

  ```bash
  curl -ik -X POST https://httpbin.$DOMAIN_TO_EXPOSE_WORKLOADS/post -d "test data"
  ```
  If successful, the call returns the `200 OK` response code.