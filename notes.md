# Notes

* glooctl install knative --install-knative-version=0.11.1
* operator_asg_name=$(aws autoscaling describe-auto-scaling-groups --region us-west-2 --query "AutoScalingGroups[?contains(Tags[?Key==\`alpha.eksctl.io/cluster-name\`].Value, \`cortex\`)]|[?contains(Tags[?Key==\`alpha.eksctl.io/nodegroup-name\`].Value, \`ng-cortex-operator\`)]" | jq -r 'first | .AutoScalingGroupName') && aws autoscaling update-auto-scaling-group --region us-west-2 --auto-scaling-group-name $operator_asg_name --max-size=3
* ./ecr_helper.sh --push-and-pull default default
* glooctl proxy url --name knative-external-proxy)
* curl $(glooctl proxy url --name knative-external-proxy)/predict -H "Host: iris-classifier.default.example.com" -H "Content-Type: application/json" -X POST -d @sample.json

## Register webhook

* mkdir -p ~/.cortex/certs && ./webhook-create-signed-cert.sh --service <dns_name> --namespace default --secret admission-webhook-example-ec2-certs
* kubectl config view --raw --flatten -o json | jq -r '.clusters[] | select(.name == "'<cluster_name>.<region>.eksctl.io'") | .cluster."certificate-authority-data"'
* copy CA_bundle for cluster and replace it in mutatingwebhook.yaml
* replace jumpbox_url in mutatingwebhook.yaml
* kubectl label namespace default admission-webhook-example=enabled
* kubectl apply -f mutatingwebhook.yaml

## Questions

* why are pods stuck in terminating for a while?
* should we move admission contoller to it's own deployment?
* image pull credentials from ecr
* check if knative automatically pulls images on new nodes before pods are scheduled, or if they have any other image-related knobs
* rolling updates
  * num replicas don't decrease
  * can we control max surge / unavailable?
    * dev update when only one replica fits on a node
    * e.g. requested 100 replicas, max nodes was 50, can you do an update after 50 replicas are running?
* logging
* metrics (response codes, e2e latency)

## TODO

* get cx deploy working
