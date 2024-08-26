package controller

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/go-openapi/runtime"
	kfpcv1alpha1 "github.com/gregsheremeta/kfpipeline-controller/api/v1alpha1"
	pcv2 "github.com/kubeflow/pipelines/backend/api/v2beta1/go_http_client/pipeline_client"
	psv2 "github.com/kubeflow/pipelines/backend/api/v2beta1/go_http_client/pipeline_client/pipeline_service"
	pucv2 "github.com/kubeflow/pipelines/backend/api/v2beta1/go_http_client/pipeline_upload_client"
	pusv2 "github.com/kubeflow/pipelines/backend/api/v2beta1/go_http_client/pipeline_upload_client/pipeline_upload_service"
	"gopkg.in/yaml.v2"
)

// Something like:
// do we already have a pipeline with this name?
//
//	if so, upload a new version
//
// else, create a new pipeline
func SyncPipeline(name string, namespace string, desiredState *kfpcv1alpha1.KFPipeline) error {

	// first, make sure we were sent something that looks like a pipeline spec
	desiredPipelineSpec, err := yamlStringToJsonString(desiredState.Spec.PipelineSpec)
	if err != nil {
		return fmt.Errorf("problem parsing the pipeline spec yaml (make sure it's valid): %v", err)
	}
	desiredDescription := desiredState.Spec.Description

	// setup clients
	// TODO deal with TLS
	pipelineConfig := &pcv2.TransportConfig{
		Host:     fmt.Sprintf("%s:%s", os.Getenv("PIPELINE_SERVICE_URL"), os.Getenv("PIPELINE_SERVICE_PORT")),
		BasePath: "/",
		Schemes:  []string{"http"},
	}
	pipelineClient := pcv2.NewHTTPClientWithConfig(nil, pipelineConfig)

	uploadConfig := &pucv2.TransportConfig{
		Host:     fmt.Sprintf("%s:%s", os.Getenv("PIPELINE_SERVICE_URL"), os.Getenv("PIPELINE_SERVICE_PORT")),
		BasePath: "/",
		Schemes:  []string{"http"},
	}
	uploadClient := pucv2.NewHTTPClientWithConfig(nil, uploadConfig)

	// check if the pipeline already exists
	currentPipelineSpec, currentDecription, exists, err := getCurrentPipeline(pipelineClient, name, namespace)
	if err != nil {
		return fmt.Errorf("problem checking for existing pipeline in KFP apiserver: %v", err)
	}

	if exists {
		fmt.Println("pipeline exists. comparing the current one in the KFP apiserver and the desired state in the CR")

		fmt.Printf("current pipeline spec: [%v]\n", currentPipelineSpec)
		fmt.Printf("desired pipeline spec: [%v]\n", desiredPipelineSpec)
		fmt.Printf("current pipeline length: %v\n", len(currentPipelineSpec))
		fmt.Printf("desired pipeline length: %v\n", len(desiredPipelineSpec))

		if currentPipelineSpec == desiredPipelineSpec {
			fmt.Println("pipeline spec is up to date. checking description")
			if currentDecription == desiredDescription {
				fmt.Println("pipeline description is up to date. Sync complete.")
				return nil
			} else {
				fmt.Println("pipeline description is out of date, updating")
				fmt.Println("not implemented")
				// updatePipelineDescription(pipelineClient, name, namespace, incoming
				return nil
			}
		} else {
			fmt.Println("pipeline is out of date, uploading new version")
			fmt.Println("not implemented")
			// uploadNewPipeline(uploadClient, name, namespace, incomingPipelineSpec, existingDecription)
			return nil
		}
	} else {
		fmt.Println("pipeline does not exist, creating new")
		err := uploadNewPipeline(uploadClient, name, namespace, desiredPipelineSpec, desiredDescription)
		if err != nil {
			return fmt.Errorf("problem uploading desired pipeline to KFP apiserver: %v", err)
		}
		return nil
	}
}

// TODO handle the case when a pipeline exists but a version does not
func getCurrentPipeline(pipelineClient *pcv2.Pipeline, name string, namespace string) (string, string, bool, error) {
	params := psv2.NewPipelineServiceGetPipelineByNameParams()
	params.SetName(name)
	params.SetNamespace(&namespace)
	response, err := pipelineClient.PipelineService.PipelineServiceGetPipelineByName(params, nil)
	if err != nil {
		// 404 will be a *pipeline_service.PipelineServiceGetPipelineByNameDefault
		if errorResponse, ok := err.(*psv2.PipelineServiceGetPipelineByNameDefault); ok {
			// 404 == grpc code 5
			// https://grpc.github.io/grpc/core/md_doc_statuscodes.html
			if errorResponse.Payload.Code == 5 {
				fmt.Println("pipeline does not exist")
				return "", "", false, nil
			}
		}
		return "", "", false, fmt.Errorf("error calling PipelineServiceGetPipelineByName: %v", err)
	}

	if response.Payload.PipelineID != "" {
		fmt.Printf("pipeline exists: [%v]\n", response.Payload)
		fmt.Printf("pipeline ID: %v\n", response.Payload.PipelineID)
		params2 := psv2.NewPipelineServiceListPipelineVersionsParams()
		params2.SetPipelineID(response.Payload.PipelineID)
		pageSize := int32(1)
		params2.SetPageSize(&pageSize)
		sortBy := "created_at desc"
		params2.SetSortBy(&sortBy)
		response2, err := pipelineClient.PipelineService.PipelineServiceListPipelineVersions(params2, nil)
		if err != nil {
			return "", "", false, fmt.Errorf("error listing pipeline versions: %v", err)
		}
		if response2.Payload.TotalSize > 0 {
			// fmt.Printf("pipeline version exists: [%v]\n", response2.Payload.PipelineVersions[0].PipelineSpec)
			pipelineSpec, err := yamlInterfaceToJsonString(response2.Payload.PipelineVersions[0].PipelineSpec)
			if err != nil {
				return "", "", false, fmt.Errorf("error converting pipeline spec to json: %v", err)
			}
			return pipelineSpec, response.Payload.Description, true, nil
		}
	}

	fmt.Printf("seems like pipeline doesn't exist: [%v]\n", response.Payload)
	fmt.Printf("id: [%v]\n", response.Payload.PipelineID)
	return "", "", false, nil
}

func uploadNewPipeline(uploadClient *pucv2.PipelineUpload, name string, namespace string, pipelineSpec string, description string) error {
	pipelineFile := runtime.NamedReader("pipeline.yaml", strings.NewReader(pipelineSpec))

	params := pusv2.NewUploadPipelineParams()
	params.SetDescription(&description)
	params.SetName(&name)
	params.SetNamespace(&namespace)
	params.SetUploadfile(pipelineFile)

	ok, err := uploadClient.PipelineUploadService.UploadPipeline(params, nil)

	// TODO add real error handling
	// TODO seems like i'm getting a 401 or something when the pipeline already exists
	fmt.Printf("ok: %v\n", ok)
	fmt.Printf("ok.Error: %v\n", ok.Error())
	fmt.Printf("err: %v\n", err)

	if err != nil {
		return fmt.Errorf("problem calling UploadPipeline: [%v], [%v]", ok.Error(), err)
	}
	return nil
}

func DeletePipeline(name string, namespace string) {
	fmt.Printf("Deleting pipeline %s in namespace %s\n", name, namespace)
	fmt.Println("not implemented")
}

// func hashPipeline(pipelineSpec string) string {
// 	pipelineSpec = strings.TrimSpace(pipelineSpec)
// 	hash := sha256.New()
// 	hash.Write([]byte(pipelineSpec))
// 	b := hash.Sum(nil)
// 	return fmt.Sprintf("%x", b)
// }

func yamlStringToJsonString(yamlString string) (string, error) {
	var data map[interface{}]interface{}
	err := yaml.Unmarshal([]byte(yamlString), &data)
	if err != nil {
		return "", fmt.Errorf("error parsing YAML: %v", err)
	}

	convertedData := convertMapInterface(data)
	jsonData, err := json.Marshal(convertedData)
	if err != nil {
		return "", fmt.Errorf("error converting to JSON: %v", err)
	}

	jsonString := string(jsonData)
	return jsonString, nil
}

func yamlInterfaceToJsonString(yamlInterface interface{}) (string, error) {
	jsonMap := convertInterface(yamlInterface)
	jsonData, err := json.Marshal(jsonMap)
	if err != nil {
		return "", fmt.Errorf("error marshaling to JSON: %v", err)
	}

	return string(jsonData), nil
}

// convert a interface{} to a map[string]interface{},
// which is required for marshaling to JSON.
func convertInterface(i interface{}) interface{} {
	switch x := i.(type) {
	case map[interface{}]interface{}:
		m2 := map[string]interface{}{}
		for k, v := range x {
			m2[k.(string)] = convertInterface(v)
		}
		return m2
	case []interface{}:
		for i, v := range x {
			x[i] = convertInterface(v)
		}
	}
	return i
}

// convert a map[interface{}]interface{} to a map[string]interface{},
// which is required for marshaling to JSON.
func convertMapInterface(input map[interface{}]interface{}) map[string]interface{} {
	output := make(map[string]interface{})
	for key, value := range input {
		// Convert key to string
		strKey := fmt.Sprintf("%v", key)

		// Convert nested maps recursively
		switch value := value.(type) {
		case map[interface{}]interface{}:
			output[strKey] = convertMapInterface(value)
		default:
			output[strKey] = value
		}
	}
	return output
}
