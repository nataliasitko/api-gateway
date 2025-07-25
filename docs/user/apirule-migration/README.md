# Migrate APIRule from Version `v1beta1` to Version `v2`
Learn how to obtain the full **spec** of an APIRule in version `v1beta1` and migrate it to version `v2`. 

## APIRule v1beta1 Deletion Timeline

The APIRule CRD in version v1beta1 has been deprecated and scheduled for deletion. Therefore, you must migrate all your APIRule CRs to version `v2`. The following timeline outlines the steps we will take to phase out version v1beta1:
1. Kyma dashboard won’t display APIRule CRs in version `v1beta1`. All APIRule CRs `v1beta1` will be fully operational from command line, and you still will be able to manage them using kubectl.
   <br>Regular channel release date: **{DD.MM.YY}**

2. You won’t be able to create APIRule CRs v1beta1 in new clusters. In existing clusters, you will still be able to create and modify APIRule CRs v1beta1.
   <br>Regular channel release date: **End of November, 2025**

3. Existing APIRule CRs v1beta1 will be reconciled, but you won’t be able to edit or delete them. You won’t be able to create APIRules v1beta1 in new and existing clusters.
   <br>Regular channel release date: **End of year 2025**

## How to Migrate APIRules to Version v2

1. To identify which APIRules must be migrated, run the following command:
    ```bash
    kubectl get apirules.gateway.kyma-project.io -A -o json | jq '.items[] | select(.metadata.annotations["gateway.kyma-project.io/original-version"] == "v1beta1") | {namespace: .metadata.namespace, name: .metadata.name}'
    ```

2. To obtain the complete **spec** with the **rules** field of an APIRule in version `v1beta1`, see [Retrieve the Complete **spec** of an APIRule in Version `v1beta1`](./01-81-retrieve-v1beta1-spec.md).


3. To migrate an APIRule from version `v1beta1` to version `v2`, see:
    - [Migrate APIRule `v1beta1` of Type **noop**, **allow** or **no_auth** to Version `v2`](./01-82-migrate-allow-noop-no_auth-v1beta1-to-v2.md)
    - [Migrate APIRule `v1beta1` of Type **jwt** to Version `v2`](./01-83-migrate-jwt-v1beta1-to-v2.md)
    - [Migrate APIRule `v1beta1` of Type **oauth2_introspection** to Version `v2`](./01-84-migrate-oauth2-v1beta1-to-v2.md)

For more information about APIRule `v2`, see also:
- [APIRule `v2` Custom Resource](../custom-resources/apirule/04-10-apirule-custom-resource.md)
- [Changes Introduced in APIRule `v2`](../custom-resources/apirule/04-70-changes-in-apirule-v2.md)