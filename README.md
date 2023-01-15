# dummy-operator
A dummy Kubernetes operator

## Installing and deploying the `dummy-operator`
You can install the operator in your local or cloud cluster by
following the simple steps from below:

- Clone this repository to your local environment:
```bash
$ git clone https://github.com/piyush1146115/dummy-operator.git`
```
- Change your current directory to the `dummy-operator`
```bash
$ cd dummy-operator
```
- Install the dummy CRD
```bash
$ make install
```

- Deploy the dummy-operator to your cluster
```bash
$ make deploy IMG=piyush1146115/dummy-operator:latest`
```

If all of the above steps were successful, you should see the dummy-operator
running in your kubernetes cluster.

## Testing

Create a dummy object in your cluster by applying the following manifest:

```yaml
apiVersion: interview.com/v1alpha1
kind: Dummy
metadata:
  name: dummy1
  namespace: default
spec:
  message: "I'm just a dummy"
```

You can also apply the manifest by applying the manifest from `/../dummy-operator/config/samples/dummy.yaml`

To get the current status of dummy objects in your `default` namespace:
```
$ kubectl get dummy -n default

NAME     SPECECHO           PODSTATUS
dummy1   I'm just a dummy   Running
```