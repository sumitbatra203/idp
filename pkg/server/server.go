package server

import (
	"atlan/idp/pkg/jobmanager"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"

	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

type Server struct {
	logger *zap.Logger
	router *mux.Router
	jm     *jobmanager.JobManager
	wg     *sync.WaitGroup
}

type JobError struct {
	Message string
	Code    int
}

func (e *JobError) Error() string {
	return fmt.Sprintf("Job Error: %s (Code: %d)", e.Message, e.Code)
}

func New(jmanager *jobmanager.JobManager) *Server {
	s := &Server{
		logger: zap.Must(zap.NewProduction()),
		jm:     jmanager,
		wg:     &sync.WaitGroup{},
	}
	s.router = s.getRoutes()
	return s
}

func (s *Server) Start() {
	s.wg.Add(1)
	var serverErr error
	go func() {
		defer s.wg.Done()
		err := http.ListenAndServe(":8181", s.router)
		if err != nil {
			serverErr = err
		}
	}()
	s.wg.Wait()
	if serverErr != nil {
		s.logger.Panic(fmt.Sprintf("error in starting server:%v", serverErr.Error()))
	}
}

func (s *Server) getRoutes() *mux.Router {
	r := mux.NewRouter()
	r.HandleFunc("/datajob", s.createJobHandler).Methods("POST")
	r.HandleFunc("/datajob", s.getJobHandler).Methods("GET")
	r.HandleFunc("/datajob/{jobID}", s.getJobHandlerById).Methods("GET")
	return r
}

func (s *Server) createJobHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	tenantName := r.Header.Get("jwt_token") //TODO:// validate signature for token and extract tenant name
	s.logger.Info("Job creation request received", zap.String("tenant", tenantName))
	err := s.jm.CreateJob(tenantName)
	if err != nil {
		s.logger.Error(fmt.Sprintf("error submitting job:%v", err.Error()), zap.String("tenant", tenantName))
		json.NewEncoder(w).Encode(&JobError{Message: err.Error(), Code: http.StatusBadRequest})
		return
	}
	s.logger.Info("Job creation request successfully", zap.String("tenant", tenantName))
	w.Write([]byte("Job creation request successfully"))
}

func (s *Server) getJobHandler(w http.ResponseWriter, r *http.Request) {
	tenantName := r.Header.Get("jwt_token") //TODO:// validate signature for token and extract tenant name
	s.logger.Info("Job get request received", zap.String("tenant", tenantName))
	paths := strings.Split(r.URL.Path, "/")
	jobID := paths[len(paths)-1]
	if jobID == "" {
		s.logger.Error("empty job id", zap.String("tenant", tenantName))
		json.NewEncoder(w).Encode(&JobError{Message: "Empty job id", Code: http.StatusBadRequest})
		return
	}
	jobResponse, err := s.jm.ReadJobs(tenantName, jobID)
	if err != nil {
		s.logger.Error(fmt.Sprintf("error reading job:%v", err.Error()), zap.String("tenant", tenantName))
		json.NewEncoder(w).Encode(&JobError{Message: err.Error(), Code: http.StatusBadRequest})
		return
	}
	fmt.Println(jobResponse)
	resBytes, err := json.Marshal(jobResponse)
	if err != nil {
		s.logger.Error(fmt.Sprintf("unable to marshal response job:%v", err.Error()), zap.String("tenant", tenantName))
		json.NewEncoder(w).Encode(&JobError{Message: err.Error(), Code: http.StatusBadRequest})
		return
	}
	_, _ = w.Write(resBytes)
}

func (s *Server) getJobHandlerById(w http.ResponseWriter, r *http.Request) {
	tenantName := r.Header.Get("jwt_token") //TODO:// validate signature for token and extract tenant name
	s.logger.Info("Job get request received", zap.String("tenant", tenantName))
	jobID := mux.Vars(r)["jobID"]
	if jobID == "" {
		s.logger.Error("empty job id", zap.String("tenant", tenantName))
		json.NewEncoder(w).Encode(&JobError{Message: "Empty job id", Code: http.StatusBadRequest})
		return
	}
	jobResponse, err := s.jm.ReadJobs(tenantName, jobID)
	if err != nil {
		s.logger.Error(fmt.Sprintf("error reading job:%v", err.Error()), zap.String("tenant", tenantName))
		json.NewEncoder(w).Encode(&JobError{Message: err.Error(), Code: http.StatusBadRequest})
		return
	}
	s.logger.Info("Job read request successfully", zap.String("tenant", tenantName))
	fmt.Println(jobResponse)
	resBytes, err := json.Marshal(jobResponse)
	if err != nil {
		s.logger.Error(fmt.Sprintf("unable to marshal response job:%v", err.Error()), zap.String("tenant", tenantName))
		json.NewEncoder(w).Encode(&JobError{Message: err.Error(), Code: http.StatusBadRequest})
		return
	}
	_, _ = w.Write(resBytes)
}
