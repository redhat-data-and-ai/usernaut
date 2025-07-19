# Deployment

1. Create the secret

```yaml
apiVersion: v1
kind: Secret
metadata:
  namespace: <namespace>
  name: usernaut-secrets
stringData:
  fivetran_key: fivetran_key
  fivetran_secret: fivetran_secret
```

1. Create the configmap with appconfig YAML

```sh
kubectl create cm usernaut-config --from-file=platformtest.yaml=./appconfig/platformtest.yaml --from-file=default.yaml=./appconfig/default.yaml -n <namespace>
```

1. In case a redis server is needed:

```yaml
apiVersion: v1
kind: Service
metadata:
  name: redis
spec:
  ports:
  - port: 6379
    protocol: TCP
    targetPort: 6379
  selector:
    run: redis
status:
  loadBalancer: {}
---
apiVersion: v1
kind: Pod
metadata:
  labels:
    run: redis
  name: redis
spec:
  containers:
  - image: redis
    name: redis
    ports:
    - containerPort: 6379
    resources:
      requests:
        cpu: 100m
        memory: 100Mi
      limits:
        cpu: 1000m
        memory: 1000Mi
  dnsPolicy: ClusterFirst
  restartPolicy: Always
status: {}
```
