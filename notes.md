# Notes

* glooctl install knative --install-knative-version=0.11.1
* operatore_asg_name=$(aws autoscaling describe-auto-scaling-groups --region us-west-2 --query "AutoScalingGroups[?contains(Tags[?Key==\`alpha.eksctl.io/cluster-name\`].Value, \`cortex\`)]|[?contains(Tags[?Key==\`alpha.eksctl.io/nodegroup-name\`].Value, \`ng-cortex-operator\`)]" | jq -r 'first | .AutoScalingGroupName') && aws autoscaling update-auto-scaling-group --region us-west-2 --auto-scaling-group-name $operatore_asg_name --max-size=3
* ./ecr_helper.sh --push-and-pull default default
* glooctl proxy url --name knative-external-proxy)
* curl $(glooctl proxy url --name knative-external-proxy)/predict -H "Host: iris-classifier.default.example.com" -H "Content-Type: application/json" -X POST -d @sample.json
