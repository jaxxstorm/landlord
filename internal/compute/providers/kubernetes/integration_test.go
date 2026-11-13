//go:build integration

package kubernetes

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"path/filepath"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/signer/v4"
	"github.com/aws/aws-sdk-go-v2/service/eks"
	ekstypes "github.com/aws/aws-sdk-go-v2/service/eks/types"
	"github.com/jaxxstorm/landlord/internal/cloud/awsconfig"
	"github.com/jaxxstorm/landlord/internal/compute"
	"github.com/jaxxstorm/landlord/internal/testutil/floci"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"
	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
)

func TestProviderIntegration(t *testing.T) {
	ctx := context.Background()
	fixture := floci.Start(t)
	fixture.SetEnvironment(t)

	awsCfg, err := awsconfig.Load(ctx, awsconfig.Options{Region: fixture.Region})
	require.NoError(t, err)
	eksClient := eks.NewFromConfig(awsCfg)
	clusterName := fixture.ResourcePrefix + "-eks"
	_, err = eksClient.CreateCluster(ctx, &eks.CreateClusterInput{
		Name:    aws.String(clusterName),
		RoleArn: aws.String(fmt.Sprintf("arn:aws:iam::%s:role/eks-test", fixture.AccessKeyID)),
		ResourcesVpcConfig: &ekstypes.VpcConfigRequest{
			SubnetIds:        []string{},
			SecurityGroupIds: []string{},
		},
		Version: aws.String("1.29"),
	})
	require.NoError(t, err)

	clusterDeleted := false
	t.Cleanup(func() {
		if clusterDeleted {
			return
		}
		cleanupCtx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
		defer cancel()
		if err := deleteCluster(cleanupCtx, eksClient, clusterName); err != nil {
			t.Errorf("delete Floci EKS cluster %q: %s", clusterName, err)
		}
	})

	cluster := waitForActiveCluster(t, ctx, eksClient, clusterName)
	endpoint := aws.ToString(cluster.Endpoint)
	require.NotEmpty(t, endpoint)
	certificateAuthority, err := base64.StdEncoding.DecodeString(aws.ToString(cluster.CertificateAuthority.Data))
	require.NoError(t, err)
	require.NotEmpty(t, certificateAuthority)

	token := flociEKSToken(t, ctx, awsCfg, fixture.Endpoint, clusterName)
	provider, err := New(Config{
		Kubeconfig: flociKubeconfig(t, endpoint, certificateAuthority, token),
	}, zaptest.NewLogger(t))
	require.NoError(t, err)

	tenantID := fixture.ResourcePrefix + "-tenant"
	tenantDestroyed := false
	spec := flociIntegrationSpec(tenantID, "initial")
	provisioned, err := provider.Provision(ctx, spec)
	require.NoError(t, err)
	require.Equal(t, compute.ProvisionStatusSuccess, provisioned.Status)
	require.NotEmpty(t, provisioned.ResourceIDs["namespace"])
	require.Equal(t, deploymentName, provisioned.ResourceIDs["deployment"])

	t.Cleanup(func() {
		if tenantDestroyed {
			return
		}
		if err := provider.Destroy(context.Background(), tenantID); err != nil {
			t.Errorf("destroy Floci Kubernetes tenant %q: %s", tenantID, err)
		}
	})

	require.Eventually(t, func() bool {
		status, err := provider.GetStatus(ctx, tenantID)
		return err == nil && status.State == compute.ComputeStateRunning
	}, 3*time.Minute, 2*time.Second)

	updated, err := provider.Update(ctx, tenantID, flociIntegrationSpec(tenantID, "updated"))
	require.NoError(t, err)
	require.Equal(t, compute.UpdateStatusSuccess, updated.Status)

	require.Eventually(t, func() bool {
		status, err := provider.GetStatus(ctx, tenantID)
		return err == nil && status.State == compute.ComputeStateRunning
	}, 3*time.Minute, 2*time.Second)

	require.NoError(t, provider.Destroy(ctx, tenantID))
	tenantDestroyed = true
	require.Eventually(t, func() bool {
		_, err := provider.GetStatus(ctx, tenantID)
		return errors.Is(err, compute.ErrTenantNotFound)
	}, 30*time.Second, time.Second)
	require.NoError(t, provider.Destroy(ctx, tenantID))

	require.NoError(t, deleteCluster(ctx, eksClient, clusterName))
	clusterDeleted = true
}

func waitForActiveCluster(t *testing.T, ctx context.Context, client *eks.Client, clusterName string) *ekstypes.Cluster {
	t.Helper()
	var cluster *ekstypes.Cluster
	require.Eventually(t, func() bool {
		result, err := client.DescribeCluster(ctx, &eks.DescribeClusterInput{Name: aws.String(clusterName)})
		if err != nil || result.Cluster == nil || result.Cluster.Status != ekstypes.ClusterStatusActive {
			return false
		}
		cluster = result.Cluster
		return true
	}, 3*time.Minute, 2*time.Second)
	return cluster
}

func flociEKSToken(t *testing.T, ctx context.Context, awsCfg aws.Config, endpoint, clusterName string) string {
	t.Helper()
	credentials, err := awsCfg.Credentials.Retrieve(ctx)
	require.NoError(t, err)

	stsURL, err := url.Parse(endpoint)
	require.NoError(t, err)
	query := stsURL.Query()
	query.Set("Action", "GetCallerIdentity")
	query.Set("Version", "2011-06-15")
	stsURL.RawQuery = query.Encode()
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, stsURL.String(), nil)
	require.NoError(t, err)
	request.Header.Set("X-K8s-Aws-Id", clusterName)

	signedURL, _, err := v4.NewSigner().PresignHTTP(ctx, credentials, request,
		"e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855", "sts", awsCfg.Region, time.Now())
	require.NoError(t, err)
	return "k8s-aws-v1." + base64.RawURLEncoding.EncodeToString([]byte(signedURL))
}

func flociKubeconfig(t *testing.T, endpoint string, certificateAuthority []byte, token string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "kubeconfig")
	err := clientcmd.WriteToFile(clientcmdapi.Config{
		Clusters: map[string]*clientcmdapi.Cluster{
			"floci": {Server: endpoint, CertificateAuthorityData: certificateAuthority},
		},
		AuthInfos: map[string]*clientcmdapi.AuthInfo{
			"floci": {Token: token},
		},
		Contexts: map[string]*clientcmdapi.Context{
			"floci": {Cluster: "floci", AuthInfo: "floci"},
		},
		CurrentContext: "floci",
	}, path)
	require.NoError(t, err)
	return path
}

func flociIntegrationSpec(tenantID, version string) *compute.TenantComputeSpec {
	return &compute.TenantComputeSpec{
		TenantID:     tenantID,
		ProviderType: "kubernetes",
		Containers: []compute.ContainerSpec{{
			Name:    "worker",
			Image:   "busybox:1.36",
			Command: []string{"sh", "-c", "sleep 3600"},
			Env:     map[string]string{"VERSION": version},
		}},
	}
}

func deleteCluster(ctx context.Context, client *eks.Client, clusterName string) error {
	_, err := client.DeleteCluster(ctx, &eks.DeleteClusterInput{Name: aws.String(clusterName)})
	return err
}
