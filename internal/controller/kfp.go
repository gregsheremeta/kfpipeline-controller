package controller

import (
	"fmt"
	"os"
	"strings"

	"github.com/go-openapi/runtime"
	kfpcv1alpha1 "github.com/gregsheremeta/kfpipeline-controller/api/v1alpha1"
	"github.com/kubeflow/pipelines/backend/api/v2beta1/go_http_client/pipeline_upload_client"
	"github.com/kubeflow/pipelines/backend/api/v2beta1/go_http_client/pipeline_upload_client/pipeline_upload_service"
)

// Something like:
// do we already have a pipeline with this name?
//
//	if so, upload a new version
//
// else, create a new pipeline
func SyncPipeline(name string, namespace string, kfPipeline *kfpcv1alpha1.KFPipeline) {

	// TODO deal with TLS
	// cfg := &pipeline_client.TransportConfig{
	// 	Host:     fmt.Sprintf("%s:%s", os.Getenv("PIPELINE_SERVICE_URL"), os.Getenv("PIPELINE_SERVICE_PORT")),
	// 	BasePath: "/",
	// 	Schemes:  []string{"http"},
	// }

	// client := pipeline_client.NewHTTPClientWithConfig(nil, cfg)

	cfg2 := &pipeline_upload_client.TransportConfig{
		Host:     fmt.Sprintf("%s:%s", os.Getenv("PIPELINE_SERVICE_URL"), os.Getenv("PIPELINE_SERVICE_PORT")),
		BasePath: "/",
		Schemes:  []string{"http"},
	}

	client2 := pipeline_upload_client.NewHTTPClientWithConfig(nil, cfg2)

	pipelineFile := runtime.NamedReader("pipeline.yaml", strings.NewReader(kfPipeline.Spec.PipelineSpec))

	params := pipeline_upload_service.NewUploadPipelineParams()
	params.SetDescription(&kfPipeline.Spec.Description)
	params.SetName(&name)
	params.SetNamespace(&namespace)
	params.SetUploadfile(pipelineFile)

	ok, err := client2.PipelineUploadService.UploadPipeline(params, nil)

	fmt.Printf("ok: %v\n", ok)
	fmt.Printf("err: %v\n", err)

}
