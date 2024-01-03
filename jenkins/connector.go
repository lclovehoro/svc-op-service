package jenkins

import (
	"context"
	"os"

	"k8s.io/klog/v2"

	"github.com/bndr/gojenkins"
)

var JenkinsClient *jenkinsClient

type jenkinsClient struct {
	*gojenkins.Jenkins
	context.Context
}

func init() {
	if err := NewJenkinsClient(); err != nil {
		klog.Errorf("Failed to connect to jenkins: %v", err)
	}
}

func NewJenkinsClient() (err error) {
	jenkinsUrl, ok := os.LookupEnv("JENKINS_URL")
	if !ok {
		jenkinsUrl = "https://jenkins-dev.test.betawm.com/"
		klog.V(3).Infof("use default jenkins url: %s", jenkinsUrl)
	}

	jenkinsUserName, ok := os.LookupEnv("JENKINS_USERNAME")
	if !ok {
		jenkinsUserName = "bt0000425"
		klog.V(3).Infof("use default jenkins username: %s", jenkinsUserName)
	}

	jenkinsToken, ok := os.LookupEnv("JENKINS_TOKEN")
	if !ok {
		jenkinsToken = "1116370325f1c09efe0301aaecfacbc739"
		klog.V(3).Infof("use default jenkins token: %s", jenkinsToken)
	}

	ctx := context.Background()
	jenkins, err := gojenkins.CreateJenkins(nil, jenkinsUrl, jenkinsUserName, jenkinsToken).Init(ctx)
	if err != nil {
		return err
	}

	JenkinsClient = &jenkinsClient{
		Jenkins: jenkins,
		Context: ctx,
	}
	return nil
}
