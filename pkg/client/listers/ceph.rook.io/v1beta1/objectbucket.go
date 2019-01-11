/*
Copyright The Kubernetes Authors.

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

// Code generated by lister-gen. DO NOT EDIT.

package v1beta1

import (
	v1beta1 "github.com/rook/rook/pkg/apis/ceph.rook.io/v1beta1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/tools/cache"
)

// ObjectBucketLister helps list ObjectBuckets.
type ObjectBucketLister interface {
	// List lists all ObjectBuckets in the indexer.
	List(selector labels.Selector) (ret []*v1beta1.ObjectBucket, err error)
	// ObjectBuckets returns an object that can list and get ObjectBuckets.
	ObjectBuckets(namespace string) ObjectBucketNamespaceLister
	ObjectBucketListerExpansion
}

// objectBucketLister implements the ObjectBucketLister interface.
type objectBucketLister struct {
	indexer cache.Indexer
}

// NewObjectBucketLister returns a new ObjectBucketLister.
func NewObjectBucketLister(indexer cache.Indexer) ObjectBucketLister {
	return &objectBucketLister{indexer: indexer}
}

// List lists all ObjectBuckets in the indexer.
func (s *objectBucketLister) List(selector labels.Selector) (ret []*v1beta1.ObjectBucket, err error) {
	err = cache.ListAll(s.indexer, selector, func(m interface{}) {
		ret = append(ret, m.(*v1beta1.ObjectBucket))
	})
	return ret, err
}

// ObjectBuckets returns an object that can list and get ObjectBuckets.
func (s *objectBucketLister) ObjectBuckets(namespace string) ObjectBucketNamespaceLister {
	return objectBucketNamespaceLister{indexer: s.indexer, namespace: namespace}
}

// ObjectBucketNamespaceLister helps list and get ObjectBuckets.
type ObjectBucketNamespaceLister interface {
	// List lists all ObjectBuckets in the indexer for a given namespace.
	List(selector labels.Selector) (ret []*v1beta1.ObjectBucket, err error)
	// Get retrieves the ObjectBucket from the indexer for a given namespace and name.
	Get(name string) (*v1beta1.ObjectBucket, error)
	ObjectBucketNamespaceListerExpansion
}

// objectBucketNamespaceLister implements the ObjectBucketNamespaceLister
// interface.
type objectBucketNamespaceLister struct {
	indexer   cache.Indexer
	namespace string
}

// List lists all ObjectBuckets in the indexer for a given namespace.
func (s objectBucketNamespaceLister) List(selector labels.Selector) (ret []*v1beta1.ObjectBucket, err error) {
	err = cache.ListAllByNamespace(s.indexer, s.namespace, selector, func(m interface{}) {
		ret = append(ret, m.(*v1beta1.ObjectBucket))
	})
	return ret, err
}

// Get retrieves the ObjectBucket from the indexer for a given namespace and name.
func (s objectBucketNamespaceLister) Get(name string) (*v1beta1.ObjectBucket, error) {
	obj, exists, err := s.indexer.GetByKey(s.namespace + "/" + name)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, errors.NewNotFound(v1beta1.Resource("objectbucket"), name)
	}
	return obj.(*v1beta1.ObjectBucket), nil
}