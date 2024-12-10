# Expose Multiple Workloads on Different Paths Using the Same Host

Learn how to expose multiple workloads on different paths by creating an APIRule and defining Services on each path separately.

> [!WARNING] Exposing a workload to the outside world is always a potential security vulnerability, so be careful. In a production environment, remember to secure the workload you expose with [OAuth2](../01-50-expose-and-secure-a-workload/01-50-expose-and-secure-workload-oauth2.md) or [JWT](../01-50-expose-and-secure-a-workload/01-52-expose-and-secure-workload-jwt.md).

## Prerequisites

* You have deployed two Services.
* You have [set up your custom domain](../01-10-setup-custom-domain-for-workload.md). Alternatively, you can use the default domain of your Kyma cluster.


## Define Multiple Services on Different Paths

<!-- tabs:start -->
#### **Kyma Dashboard**

1. Go to **Discovery and Network > APIRules** and select **Create**. 
2. Add `multiple-services` as a name.
3. In the Gateway section, provide the Gateway's namespace and name. If you use a Kyma domain, use the `kyma-system` namespace and `kyma-gateway` name.
4. In the **Host** field, enter `multiple-services.{DOMAIN_TO_EXPORT_WORKLOADS}`. Replace the placeholder with the name of your domain.
5. In the **Rules > Rule** section, add the configuration details for one of your Services.
6. Create one more **Rule** and add the configuration details of the other Service.
7. To create the APIRule, select **Create**.


#### **kubectl**
1. Export the name, port, and path of one of your Service.
  
  ```bash
  export SERVICE_NAME1={SERVICE_NAME}
  export SERVICE_PORT1={SERVICE_PORT}
  export SERVICE_PATH1={SERVICE_PATH}
  ```
1. Export the name, port, and path of the other Service.
  
  ```bash
  export SERVICE_NAME2={SERVICE_NAME}
  export SERVICE_PORT2={SERVICE_PORT}
  export SERVICE_PATH2={SERVICE_PATH}
  ```

3. Depending on whether you use your custom domain or a Kyma domain, export the necessary values as environment variables:
  
  <!-- tabs:start -->
  #### **Custom Domain**
      
  ```bash
  export DOMAIN_TO_EXPOSE_WORKLOADS={DOMAIN_NAME}
  export GATEWAY=$NAMESPACE/httpbin-gateway
  ```
  #### **Kyma Domain**

  ```bash
  export DOMAIN_TO_EXPOSE_WORKLOADS={KYMA_DOMAIN_NAME}
  export GATEWAY=kyma-system/kyma-gateway
  ```
  <!-- tabs:end --> 

4. To expose the Services, create the following APIRule. You can adjust the **rules** configuration according to your needs.

    ```bash
    cat <<EOF | kubectl apply -f -
    apiVersion: gateway.kyma-project.io/v1beta1
    kind: APIRule
    metadata:
      name: multiple-services
      namespace: $NAMESPACE
      labels:
        app: multiple-services
        example: multiple-services
    spec:
      host: multiple-services.$DOMAIN_TO_EXPOSE_WORKLOADS
      gateway: $GATEWAY
      rules:
      - path: $SERVICE_PATH1
        methods: ["GET"]
        accessStrategies:
          - handler: no_auth
        service:
          name: $SERVICE_NAME1
          port: $SERVICE_PORT1
      - path: $SERVICE_PATH2
        methods: ["GET"]
        accessStrategies:
          - handler: no_auth
        service:
          name: $SERVICE_NAME2
          port: $SERVICE_PORT2
    EOF
    ```
<!-- tabs:end -->

## Result
You have exposed the Services. To call the endpoints, send `GET` requests:

  ```bash
  curl -ik -X GET https://multiple-services.{DOMAIN_TO_EXPOSE_WORKLOADS}/{PATH1}

  curl -ik -X GET https://multiple-services.{DOMAIN_TO_EXPOSE_WORKLOADS}/{PATH2} 
  ```
If successful, the calls return the `200 OK` response code.