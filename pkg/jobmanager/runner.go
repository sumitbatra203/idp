package jobmanager

import (
	"context"
	"fmt"
	"strings"

	"github.com/goccy/kubejob"
	"github.com/lithammer/shortuuid/v4"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	v1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/client-go/rest"
)

var (
	containerName = "idp-job-runner"
	imageName     = "nginx"
	jobPrefix     = "idp-"
)

type JobManager struct {
	logger     *zap.Logger
	jobBuilder *kubejob.JobBuilder
}

type JobResponse struct {
	Name   string `json:"name"`
	Id     string `json:"Id"`
	Status string `json:"Status"`
}

type JobStatus int

const (
	SUBMITTED JobStatus = iota
	RUNNING
	FAILED
	COMPLETED
)

func (js JobStatus) String() string {
	return []string{"Submit", "Running", "Failed", "Completed"}[js]
}

func NewJobManager(config *rest.Config, namespace string) *JobManager {
	return &JobManager{
		logger:     zap.Must(zap.NewProduction()),
		jobBuilder: kubejob.NewJobBuilder(config, namespace),
	}
}

func (m *JobManager) CreateJob(tenantName string) error {
	shortUid := shortuuid.New()
	jobTemplate := m.getJobTemplate(shortUid, []string{"ls"})
	job, err := m.jobBuilder.BuildWithJob(jobTemplate)
	if err != nil {
		m.logger.Error(fmt.Sprintf("Error in building job :%v", err.Error()), zap.String("tenant", tenantName))
		return errors.Wrap(err, "error in building job")
	}
	m.logger.Info(fmt.Sprintf("job submitted with name:%s", jobTemplate.Name), zap.String("tenant", tenantName))
	//TODO://submit job to queue and db for tracking
	err = job.Run(context.Background())
	if err != nil {
		m.logger.Error(fmt.Sprintf("Error in running job :%v", err.Error()), zap.String("tenant", tenantName))
		return errors.Wrap(err, "error in running job")
	}
	m.logger.Info(fmt.Sprintf("job run successfully with name:%s", jobTemplate.Name), zap.String("tenant", tenantName))
	return nil
}

func (m *JobManager) getJobTemplate(shortUID string, commands []string) *v1.Job {
	return &v1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name: fmt.Sprintf("%s%s", jobPrefix, strings.ToLower(shortUID)),
		},
		Spec: v1.JobSpec{
			Template: corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:    containerName,
							Image:   imageName,
							Command: append([]string{}, commands...),
						},
					},
				},
			},
		},
	}
}

func (m *JobManager) ReadJobs(tenantName string, jobId string) ([]JobResponse, error) {
	m.logger.Info("get request received with name", zap.String("tenant", tenantName))
	//TODO://read job status from DB
	return []JobResponse{
		{
			Name:   "test-job-1",
			Id:     "2025022210001",
			Status: JobStatus(0).String(),
		},
		{
			Name:   "test-job-2",
			Id:     "2025022210002",
			Status: JobStatus(1).String(),
		},
		{
			Name:   "test-job-3",
			Id:     "2025022210003",
			Status: JobStatus(2).String(),
		},
	}, nil
}
