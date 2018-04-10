package main

import (
	"flag"
	"fmt"
	"github.com/golang/glog"
	"math/rand"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/apiserver/pkg/util/logs"
	restclient "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	//kapiv1 "k8s.io/kubernetes/pkg/apis/core/v1"

	//projectv1client "github.com/openshift/client-go/project/clientset/versioned/typed/project/v1"
	imageapiv1 "github.com/openshift/api/image/v1"
	imagev1client "github.com/openshift/client-go/image/clientset/versioned/typed/image/v1"
)

const (
	createFailure                 = "CreateFailure"
	deleteFailure                 = "DeleteFailure"
	getFailure                    = "GetFailure"
	importFailure                 = "ImportFailure"
	unknownFailure                = "UnknownFailure"
	artificialFailure             = "ArtificialFailure"
	defaultTestImageRepository    = "docker.io/openshift/jenkins-2-centos7"
	defaultTestImageRepositoryTag = "latest"
	defaultTestImageStream        = "importtest"
	defaultTestImageStreamTag     = "tag"
	defaultFailurePercent         = 0
	defaultInterval               = 60
)

var (
	testImageRepository, testImageRepositoryTag, testImageStream, testImageStreamTag *string
	interval, failurePercent                                                         *int
)

func main() {
	logs.InitLogs()
	addr := flag.String("listen-address", ":8443", "The address to listen on for HTTPS requests.")
	testImageRepository = flag.String("import-repository", defaultTestImageRepository, "The docker image repository to attempt to import")
	testImageRepositoryTag = flag.String("import-repository-tag", defaultTestImageRepositoryTag, "The docker image repository tag to attempt to import")
	testImageStream = flag.String("imagestream", defaultTestImageStream, "The imagestream name to create/import into")
	testImageStreamTag = flag.String("imagestreamtag", defaultTestImageStreamTag, "The imagestream tag name to create/import into")
	failurePercent = flag.Int("failure-percent", defaultFailurePercent, "Percent of test runs to forcibly fail")
	interval = flag.Int("interval", defaultInterval, "Interval between test runs, in seconds")

	flag.Parse()

	http.HandleFunc("/healthz", handleHealthz)
	http.Handle("/metrics", prometheus.Handler())

	imageStreamImportLastRun := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "imagestream_import_last_run",
			Help: "Indicates the last time the ImageStream import test ran and the result.",
		},
		[]string{"result", "reason"},
	)
	prometheus.MustRegister(imageStreamImportLastRun)

	rand.Seed(time.Now().UTC().UnixNano())

	//go http.ListenAndServe(*addr, nil)
	go http.ListenAndServeTLS(*addr, "/etc/tls-volume/tls.crt", "/etc/tls-volume/tls.key", nil)

	go runImageStreamImport(imageStreamImportLastRun, time.Duration(*interval)*time.Second)

	select {}
}

func handleHealthz(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "ok")
}

// Creates a rest config object that is used for other client calls.
func getRestConfig() *restclient.Config {

	clientConfig, err := restclient.InClusterConfig()
	if err != nil {
		panic(err)
	}

	return clientConfig
}

func runImageStreamImport(lastran *prometheus.GaugeVec, interval time.Duration) {
	imageclient, err := imagev1client.NewForConfig(getRestConfig())

	imageStream := &imageapiv1.ImageStream{
		ObjectMeta: metav1.ObjectMeta{Name: *testImageStream},
		Spec: imageapiv1.ImageStreamSpec{
			DockerImageRepository: "",
			Tags: []imageapiv1.TagReference{
				{
					Name: *testImageStreamTag,
					From: &corev1.ObjectReference{
						Kind: "DockerImage",
						Name: *testImageRepository + ":" + *testImageRepositoryTag,
					},
				},
			},
		},
	}

	kc := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		clientcmd.NewDefaultClientConfigLoadingRules(),
		&clientcmd.ConfigOverrides{},
	)
	ns, _, err := kc.Namespace()
	if err != nil {
		lastran.WithLabelValues("failed", createFailure).SetToCurrentTime()
		glog.Errorf("imageStream creation failed while retrieving client namespace with error: %v", err)
		return
	}
	first := true
	for {
		if !first {
			time.Sleep(interval)
		} else {
			first = false
		}
		glog.V(2).Infoln("Running imagestream import smoke test")

		if *failurePercent > 0 && rand.Int31n(101) < int32(*failurePercent) {
			glog.Infoln("Forced failure")
			lastran.WithLabelValues("failed", artificialFailure).SetToCurrentTime()
			continue
		}
		// Start with a clean slate.
		err = imageclient.ImageStreams(ns).Delete(imageStream.Name, nil)
		if err != nil && !apierrors.IsNotFound(err) {
			lastran.WithLabelValues("failed", deleteFailure).SetToCurrentTime()
			glog.Errorf("imagestream deletion failed with error=%v", err)
			continue
		}
		_, err = imageclient.ImageStreams(ns).Create(imageStream)
		if err != nil {
			lastran.WithLabelValues("failed", createFailure).SetToCurrentTime()
			glog.Errorf("imagestream creation failed with error: %v", err)
			continue
		}

		err = wait.Poll(time.Second, 60*time.Second, func() (bool, error) {
			_, err := imageclient.ImageStreamTags(ns).Get(*testImageStream+":"+*testImageStreamTag, metav1.GetOptions{})
			if err == nil {
				return true, nil
			}
			if apierrors.IsNotFound(err) {
				return false, nil
			}
			lastran.WithLabelValues("failed", getFailure).SetToCurrentTime()
			return false, err
		})
		if err == wait.ErrWaitTimeout {
			lastran.WithLabelValues("failed", importFailure).SetToCurrentTime()
			glog.Errorf("imagestream import timed out: %v, current imagestream is: %v", err, imageStream)
			continue
		} else if err != nil {
			lastran.WithLabelValues("failed", unknownFailure).SetToCurrentTime()
			glog.Errorf("imagestream import failed: %v, current imagestream is: %v", err, imageStream)
			continue
		}
		glog.V(2).Infof("Import successful")
		lastran.WithLabelValues("success", "").SetToCurrentTime()
	}

}
