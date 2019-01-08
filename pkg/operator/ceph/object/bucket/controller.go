/*
Copyright 2016 The Rook Authors. All rights reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package objectbucket

import (
	"fmt"
	"reflect"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"

	"github.com/coreos/pkg/capnslog"
	opkit "github.com/rook/operator-kit"
	"k8s.io/api/core/v1"
	apiextensionsv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/selection"
	v12 "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/tools/cache"

	"k8s.io/client-go/informers"

	cephv1beta1 "github.com/rook/rook/pkg/apis/ceph.rook.io/v1beta1"
	"github.com/rook/rook/pkg/client/informers/externalversions"
	obListerv1beta1 "github.com/rook/rook/pkg/client/listers/ceph.rook.io/v1beta1"
	"github.com/rook/rook/pkg/clusterd"
)

const rookCephPrefix string = "rook-ceph-object-bucket-"

// ObjectBucketResource represent the object store user custom resource for the watcher
var ObjectBucketResource = opkit.CustomResource{
	Name:    "cephobjectbucket",
	Plural:  "cephobjectbucket",
	Group:   cephv1beta1.CustomResourceGroup,
	Version: cephv1beta1.Version,
	Scope:   apiextensionsv1beta1.NamespaceScoped,
	Kind:    reflect.TypeOf(cephv1beta1.ObjectBucket{}).Name(),
}

var logger = capnslog.NewPackageLogger("github.com/rook/rook", "op-object")

// Controller encapsulates the object bucket controller event handling
type Controller struct {
	context      *clusterd.Context
	ownerRef     metav1.OwnerReference
	bucketLister obListerv1beta1.ObjectBucketLister
	secretLister v12.SecretLister
}

// NewObjectBucketController create controller for watching object bucket custom resources
// A shared indexer is created to reduce calls to the API server.
func NewObjectBucketController(context *clusterd.Context, ownerRef metav1.OwnerReference) *Controller {
	externalInformerFactory := externalversions.NewSharedInformerFactory(context.RookClientset, 0)
	internalInformerFactory := informers.NewSharedInformerFactory(context.Clientset, 0)
	return &Controller{
		context:      context,
		ownerRef:     ownerRef,
		bucketLister: externalInformerFactory.Ceph().V1beta1().ObjectBuckets().Lister(),
		secretLister: internalInformerFactory.Core().V1().Secrets().Lister(),
	}

}

// StartWatch watches ObjectBucket custom resources and acts on API events
func (c *Controller) StartWatch(stopCh <-chan struct{}) {
	resourceHandlerFuncs := cache.ResourceEventHandlerFuncs{
		AddFunc:    c.onAdd,
		UpdateFunc: c.onUpdate,
		DeleteFunc: c.onDelete,
	}

	logger.Infof("start watching object bucket resources")
	watcher := opkit.NewWatcher(ObjectBucketResource, "", resourceHandlerFuncs, c.context.RookClientset.CephV1beta1().RESTClient())

	go watcher.Watch(&cephv1beta1.ObjectBucket{}, stopCh)
	return
}

func (c *Controller) onAdd(obj interface{}) {
	objectBucket, err := c.getObjectBucketResource(obj)
	if err != nil {
		logger.Errorf("failed to get objectbucket resource: %v", err)
		return
	}
	c.handleAdd(objectBucket)
}

func (c *Controller) onUpdate(oldObj, newObj interface{}) {
	// TODO
}

func (c *Controller) onDelete(obj interface{}) {
	// TODO
}

func (c *Controller) handleAdd(ob *cephv1beta1.ObjectBucket) {
	// get bucket owner
	user, err := c.getObjectBucketUserFromBucket(ob)
	if err != nil {
		logger.Errorf("could not get bucket owner: %v", err)
		return
	}
	// get secret by labels.user
	secret, err := c.getObjectBucketUserSecret(user, ob.Namespace)
	if err != nil {
		logger.Errorf("could not get secret via object user: %v", err)
		return
	}
	// rgw create bucket with user/creds

	// create configMap
}

func (c *Controller) handleUpdate(ob *cephv1beta1.ObjectBucket) {
	// TODO
}

func (c *Controller) handleDelete(ob *cephv1beta1.ObjectBucket) {
	// TODO
}

func (c *Controller) getObjectBucketResource(obj interface{}) (*cephv1beta1.ObjectBucket, error) {
	var ok bool
	objectBucket, ok := obj.(cephv1beta1.ObjectBucket)
	if ok {
		return objectBucket.DeepCopy(), nil
	}
	return nil, fmt.Errorf("obj does not match ObjectBucket type")
}

func (c *Controller) getObjectBucketUserFromBucket(ob *cephv1beta1.ObjectBucket) (string, error) {
	user := ob.Spec.ObjectUser
	if user == "" {
		return "", fmt.Errorf("ObjectUser is empty, required")
	}
	return user, nil
}

func (c *Controller) getObjectBucketUserSecret(user, namespace string) (*v1.Secret, error) {
	req, err := labels.NewRequirement("user", selection.Equals, []string{user})
	if err != nil {
		return nil, fmt.Errorf("error creating label requirement: %v", err)
	}
	secretList, err := c.secretLister.Secrets(namespace).List(labels.NewSelector().Add(*req))
	if err != nil {
		return nil, fmt.Errorf("could not list user secrets by 'user' label: %v", err)
	}
	if n := len(secretList); n != 1 {
		// unexpected edge case. A ceph object user can only have 1 set of access keys.  If < 1 <  secrets exist we
		// will not attempt to determine which one is legitimate.
		return nil, fmt.Errorf("expected to find 1 secret for namespace/user %s/%s, got %d", namespace, user, n)
	}
	return secretList[0], nil
}

// TODO
func (c *Controller) newBucketConfigResource(ob *cephv1beta1.ObjectBucket) (*v1.ConfigMap, error) {
	config := &v1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name: rookCephPrefix + ob.Name,
		},
		Data: map[string]string{
			"BUCKET_HOST": "",
			"BUCKET_PORT": "",
			"BUCKET_NAME": "",
			"BUCKET_SSL":  "",
		},
	}
	return c.context.Clientset.CoreV1().ConfigMaps(ob.Namespace).Create(config)
}

func (c *Controller) createCephBucket(id, secret, bucketName string) {
	// s3Client := newS3ClientFromCreds(id, secret)
	// output, err := s3Client.CreateBucket(&s3.CreateBucketInput{
	//
	// })
}

func newS3ClientFromCreds(id, secret string) *s3.S3 {
	sess := session.Must(session.NewSession(&aws.Config{
		Credentials: credentials.NewStaticCredentials(id, secret, ""),
	}))
	return s3.New(sess, aws.NewConfig())
}
