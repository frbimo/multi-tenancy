package blockuseofnodeportservices

import (
	"context"
	"fmt"
	deploymentutil "sigs.k8s.io/multi-tenancy/benchmarks/kubectl-mtb/test/utils/resources/deployment"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/uuid"
	imageutils "k8s.io/kubernetes/test/utils/image"
	"sigs.k8s.io/multi-tenancy/benchmarks/kubectl-mtb/bundle/box"
	"sigs.k8s.io/multi-tenancy/benchmarks/kubectl-mtb/pkg/benchmark"
	"sigs.k8s.io/multi-tenancy/benchmarks/kubectl-mtb/test"
	"sigs.k8s.io/multi-tenancy/benchmarks/kubectl-mtb/test/utils"
	serviceutil "sigs.k8s.io/multi-tenancy/benchmarks/kubectl-mtb/test/utils/resources/service"
	"sigs.k8s.io/multi-tenancy/benchmarks/kubectl-mtb/types"
)

var b = &benchmark.Benchmark{

	PreRun: func(options types.RunOptions) error {

		resources := []utils.GroupResource{
			{
				APIGroup: "",
				APIResource: metav1.APIResource{
					Name: "services",
				},
			},
			{
				APIGroup: "apps",
				APIResource: metav1.APIResource{
					Name: "deployments",
				},
			},
		}

		for _, resource := range resources {
			access, msg, err := utils.RunAccessCheck(options.TClient, options.TenantNamespace, resource, "create")
			if err != nil {
				options.Logger.Debug(err.Error())
				return err
			}
			if !access {
				return fmt.Errorf(msg)
			}
		}

		return nil
	},

	Run: func(options types.RunOptions) error {

		podLabels := map[string]string{"test": "multi"}
		deploymentName := "deployment-" + string(uuid.NewUUID())
		imageName := "image-" + string(uuid.NewUUID())
		deploymentSpec := &deploymentutil.DeploymentSpec{DeploymentName: deploymentName, Replicas: 1, PodLabels: podLabels, ImageName: imageName, Image: imageutils.GetE2EImage(imageutils.Nginx), StrategyType: "Recreate"}
		err := deploymentSpec.SetDefaults()
		if err != nil {
			options.Logger.Debug(err.Error())
			return err
		}

		deployment := deploymentSpec.GetDeployment()
		_, err = options.TClient.AppsV1().Deployments(options.TenantNamespace).Create(context.TODO(), deployment, metav1.CreateOptions{DryRun: []string{metav1.DryRunAll}})
		if err != nil {
			options.Logger.Debug(err.Error())
			return err
		}

		svcSpec := &serviceutil.ServiceConfig{Type: "NodePort", Selector: podLabels}
		svc := svcSpec.CreateServiceSpec()
		_, err = options.TClient.CoreV1().Services(options.TenantNamespace).Create(context.TODO(), svc, metav1.CreateOptions{DryRun: []string{metav1.DryRunAll}})

		if err == nil {
			return fmt.Errorf("Tenant must be unable to create service of type NodePort")
		}
		options.Logger.Debug("Test Passed: ", err.Error())
		return nil
	},
}

func init() {
	// Get the []byte representation of a file, or an error if it doesn't exist:
	err := b.ReadConfig(box.Get("block_use_of_nodeport_services/config.yaml"))
	if err != nil {
		fmt.Println(err.Error())
	}

	test.BenchmarkSuite.Add(b)
}
