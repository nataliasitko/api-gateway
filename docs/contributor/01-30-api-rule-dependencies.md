# Dependencies of APIRules

## Istio

APIRules require Istio to be installed on the cluster because APIRule Controller creates the `VirtualService`, `AuthorizationPolicy`, and `RequestAuthentication` custom resources, which are provided by Istio.

## Ory Oathkeeper

> **CAUTION:** Ory Oathkeeper has been deprecated. This dependency will change in the future.

To use APIRules, both Ory Oathkeeper and Ory Oathkeeper Maester must be installed on the cluster. This is required because APIRule Controller creates the Rule custom resource when an APIRule has defined an access strategy other than `allow`.

You can install Ory Oathkeeper by installing the API Gateway module.