# Notes

* glooctl install knative --install-knative-version=0.11.1
* operator_asg_name=$(aws autoscaling describe-auto-scaling-groups --region us-west-2 --query "AutoScalingGroups[?contains(Tags[?Key==\`alpha.eksctl.io/cluster-name\`].Value, \`cortex\`)]|[?contains(Tags[?Key==\`alpha.eksctl.io/nodegroup-name\`].Value, \`ng-cortex-operator\`)]" | jq -r 'first | .AutoScalingGroupName') && aws autoscaling update-auto-scaling-group --region us-west-2 --auto-scaling-group-name $operator_asg_name --max-size=3
* ./ecr_helper.sh --push-and-pull default default
* glooctl proxy url --name knative-external-proxy)
* curl $(glooctl proxy url --name knative-external-proxy)/predict -H "Host: iris-classifier.default.example.com" -H "Content-Type: application/json" -X POST -d @sample.json


# Register webhook
* ./webhook-create-signed-cert.sh --service <dns_name>  --namespace default --secret admission-webhook-example-ec2-certs (generate TLS certificates so that the api-server accepts requests)
* kubectl config view --raw --flatten -o json | jq -r '.clusters[] | select(.name == "'<cluster_name>.<region>.eksctl.io'") | .cluster."certificate-authority-data"' (get CA_bundle for cluster and replace it in mutatingwebhook.yaml)
* replace jumpbox_url in mutatingwebhook.yaml
* kubectl label namespace default admission-webhook-example=enabled
* kubectl apply -f mutatingwebhook.yaml
