# Notes

* glooctl install knative --install-knative-version=0.11.1
* ./ecr_helper.sh --push-and-pull
* glooctl proxy url --name knative-external-proxy)
* curl $(glooctl proxy url --name knative-external-proxy)/predict -H "Host: iris-classifier.default.example.com" -H "Content-Type: application/json" -X POST -d @sample.json
