package db

import "k8s.io/klog/v2"

var (
	StatementSamplesClient = statementSamplesClient{}
)

type statementSamplesClient struct{}

func (p *statementSamplesClient) SubmitEvents() {
	klog.V(5).Infof("SubmitEvents")
	return
}
