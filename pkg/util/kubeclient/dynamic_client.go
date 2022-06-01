/*
Copyright 2020 The SuperEdge Authors.
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

package kubeclient

import (
	"bytes"
	"context"
	"io"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	kuberuntime "k8s.io/apimachinery/pkg/runtime"
	rutimescheme "k8s.io/apimachinery/pkg/runtime/schema"
	yamlutil "k8s.io/apimachinery/pkg/util/yaml"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/restmapper"
	"strings"
)

/*
1. 推荐使用dynamicClient创建和删除资源
2. 需要替换的变量不在matadata中，可以不替换直接删除
*/
func CreateOrDeleteResourceWithFile(client kubernetes.Interface, dynamicClient dynamic.Interface, strtmpl string, option map[string]interface{}, create bool) error {

	var f io.Reader
	var err error
	data, err := ParseString(strtmpl, option)
	if err != nil {
		return err
	}
	f = bytes.NewBuffer(data)

	d := yamlutil.NewYAMLOrJSONDecoder(f, 4096)
	dc := client.Discovery()
	restMapperRes, err := restmapper.GetAPIGroupResources(dc)
	if err != nil {
		return err
	}

	restMapper := restmapper.NewDiscoveryRESTMapper(restMapperRes)

	for {
		ext := kuberuntime.RawExtension{}
		if err := d.Decode(&ext); err != nil {
			if err == io.EOF {
				break
			}
			return err
		}

		obj, gvk, err := unstructured.UnstructuredJSONScheme.Decode(ext.Raw, nil, nil)
		if err != nil {
			return err
		}

		unstructuredMap, err := kuberuntime.DefaultUnstructuredConverter.ToUnstructured(obj)
		if err != nil {
			return err
		}

		unstructuredObj := &unstructured.Unstructured{Object: unstructuredMap}
		mapping, err := restMapper.RESTMapping(gvk.GroupKind(), gvk.Version)
		if err != nil {
			return err
		}

		if dynamicClient != nil {

			if create {
				//使用dynamicClient创建资源
				if mapping.Scope.Name() == meta.RESTScopeNameNamespace {
					_, err = dynamicClient.Resource(mapping.Resource).Namespace(unstructuredObj.GetNamespace()).Create(context.Background(), unstructuredObj, metav1.CreateOptions{})
					if err != nil {
						//存在则更新
						if apierrors.IsAlreadyExists(err) {
							_, err = dynamicClient.Resource(mapping.Resource).Namespace(unstructuredObj.GetNamespace()).Update(context.Background(), unstructuredObj, metav1.UpdateOptions{})
							if err != nil {
								return err
							}
						} else {
							return err
						}
					}
				} else {
					_, err := dynamicClient.Resource(mapping.Resource).Create(context.Background(), unstructuredObj, metav1.CreateOptions{})
					if err != nil {
						//存在则更新
						if apierrors.IsAlreadyExists(err) {
							_, err = dynamicClient.Resource(mapping.Resource).Update(context.Background(), unstructuredObj, metav1.UpdateOptions{})
							if err != nil {
								return err
							}
						} else {
							return err
						}
					}
				}
			} else {
				//使用dynamicClient删除资源
				if mapping.Scope.Name() == meta.RESTScopeNameNamespace {
					err = dynamicClient.Resource(mapping.Resource).Namespace(unstructuredObj.GetNamespace()).Delete(context.Background(), unstructuredObj.GetName(), metav1.DeleteOptions{})
					if err != nil && !apierrors.IsNotFound(err) {
						return err
					}
				} else {
					err = dynamicClient.Resource(mapping.Resource).Delete(context.Background(), unstructuredObj.GetName(), metav1.DeleteOptions{})
					if err != nil && !apierrors.IsNotFound(err) {
						return err
					}
				}

			}
		} else {
			groupVersion := rutimescheme.GroupVersion{
				Group:   mapping.GroupVersionKind.Group,
				Version: mapping.GroupVersionKind.Version,
			}

			register := scheme.Scheme.IsVersionRegistered(groupVersion)
			if !register {
				metav1.AddToGroupVersion(scheme.Scheme, groupVersion)
			}

			url := []string{}
			if len(groupVersion.Group) == 0 {
				url = append(url, "api")
			} else {
				url = append(url, "apis", groupVersion.Group)
			}
			url = append(url, groupVersion.Version)
			if create {
				var result rest.Result

				//使用restClient创建资源
				outBytes, err := kuberuntime.Encode(unstructured.UnstructuredJSONScheme, unstructuredObj)
				if err != nil {
					return err
				}
				if mapping.Scope.Name() == meta.RESTScopeNameNamespace {
					result = client.Discovery().RESTClient().Post().SpecificallyVersionedParams(&metav1.CreateOptions{}, scheme.ParameterCodec, groupVersion).AbsPath(strings.Join(url, "/")).Resource(mapping.Resource.Resource).Namespace(unstructuredObj.GetNamespace()).Body(outBytes).Do(context.Background())
					if result.Error() != nil {
						//存在则更新
						if apierrors.IsAlreadyExists(result.Error()) {
							result = client.Discovery().RESTClient().Put().SpecificallyVersionedParams(&metav1.CreateOptions{}, scheme.ParameterCodec, groupVersion).AbsPath(strings.Join(url, "/")).Resource(mapping.Resource.Resource).Namespace(unstructuredObj.GetNamespace()).Body(outBytes).Do(context.Background())
							if result.Error() != nil {
								return result.Error()
							}
						} else {
							return result.Error()
						}
					}

				} else {
					result = client.Discovery().RESTClient().Post().SpecificallyVersionedParams(&metav1.CreateOptions{}, scheme.ParameterCodec, groupVersion).AbsPath(strings.Join(url, "/")).Resource(mapping.Resource.Resource).Body(outBytes).Do(context.Background())
					//存在则更新
					if apierrors.IsAlreadyExists(result.Error()) {
						result = client.Discovery().RESTClient().Put().SpecificallyVersionedParams(&metav1.CreateOptions{}, scheme.ParameterCodec, groupVersion).AbsPath(strings.Join(url, "/")).Resource(mapping.Resource.Resource).Name(unstructuredObj.GetName()).Body(outBytes).Do(context.Background())
						if result.Error() != nil {
							return result.Error()
						}
					} else {
						return result.Error()
					}
				}
			} else {
				var result rest.Result
				//使用restClient删除资源
				if mapping.Scope.Name() == meta.RESTScopeNameNamespace {
					result = client.Discovery().RESTClient().Delete().SpecificallyVersionedParams(&metav1.DeleteOptions{}, scheme.ParameterCodec, groupVersion).AbsPath(strings.Join(url, "/")).Resource(mapping.Resource.Resource).Namespace(unstructuredObj.GetNamespace()).Name(unstructuredObj.GetName()).Do(context.Background())
				} else {
					result = client.Discovery().RESTClient().Delete().SpecificallyVersionedParams(&metav1.DeleteOptions{}, scheme.ParameterCodec, groupVersion).AbsPath(strings.Join(url, "/")).Resource(mapping.Resource.Resource).Name(unstructuredObj.GetName()).Do(context.Background())
				}

				if result.Error() != nil && !apierrors.IsNotFound(result.Error()) {
					return err
				}
			}
		}
	}
	return nil
}
