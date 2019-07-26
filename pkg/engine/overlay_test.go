package engine

import (
	"encoding/json"
	"reflect"
	"testing"

	jsonpatch "github.com/evanphx/json-patch"
	"gotest.tools/assert"
)

func compareJSONAsMap(t *testing.T, expected, actual []byte) {
	var expectedMap, actualMap map[string]interface{}
	assert.NilError(t, json.Unmarshal(expected, &expectedMap))
	assert.NilError(t, json.Unmarshal(actual, &actualMap))
	assert.Assert(t, reflect.DeepEqual(expectedMap, actualMap))
}

func TestApplyOverlay_NestedListWithAnchor(t *testing.T) {
	resourceRaw := []byte(`
	 {  
		"apiVersion":"v1",
		"kind":"Endpoints",
		"metadata":{  
		   "name":"test-endpoint",
		   "labels":{  
			  "label":"test"
		   }
		},
		"subsets":[  
		   {  
			  "addresses":[  
				 {  
					"ip":"192.168.10.171"
				 }
			  ],
			  "ports":[  
				 {  
					"name":"secure-connection",
					"port":443,
					"protocol":"TCP"
				 }
			  ]
		   }
		]
	 }`)

	overlayRaw := []byte(`
	 {  
		"subsets":[  
		   {  
			  "ports":[  
				 {  
					"(name)":"secure-connection",
					"port":444,
					"protocol":"UDP"
				 }
			  ]
		   }
		]
	 }`)

	var resource, overlay interface{}

	json.Unmarshal(resourceRaw, &resource)
	json.Unmarshal(overlayRaw, &overlay)

	patches, err := applyOverlay(resource, overlay, "/")
	assert.NilError(t, err)
	assert.Assert(t, patches != nil)

	patch := JoinPatches(patches)
	decoded, err := jsonpatch.DecodePatch(patch)
	assert.NilError(t, err)
	assert.Assert(t, decoded != nil)

	patched, err := decoded.Apply(resourceRaw)
	assert.NilError(t, err)
	assert.Assert(t, patched != nil)

	expectedResult := []byte(`
	 {  
		"apiVersion":"v1",
		"kind":"Endpoints",
		"metadata":{  
		   "name":"test-endpoint",
		   "labels":{  
			  "label":"test"
		   }
		},
		"subsets":[  
		   {  
			  "addresses":[  
				 {  
					"ip":"192.168.10.171"
				 }
			  ],
			  "ports":[  
				 {  
					"name":"secure-connection",
					"port":444.000000,
					"protocol":"UDP"
				 }
			  ]
		   }
		]
	 }`)

	compareJSONAsMap(t, expectedResult, patched)
}

func TestApplyOverlay_InsertIntoArray(t *testing.T) {
	resourceRaw := []byte(`
	 {  
		"apiVersion":"v1",
		"kind":"Endpoints",
		"metadata":{  
		   "name":"test-endpoint",
		   "labels":{  
			  "label":"test"
		   }
		},
		"subsets":[  
		   {  
			  "addresses":[  
				 {  
					"ip":"192.168.10.171"
				 }
			  ],
			  "ports":[  
				 {  
					"name":"secure-connection",
					"port":443,
					"protocol":"TCP"
				 }
			  ]
		   }
		]
	 }`)
	overlayRaw := []byte(`
	 {  
		"subsets":[  
		   {  
			  "addresses":[  
				 {  
					"ip":"192.168.10.172"
				 },
				 {  
					"ip":"192.168.10.173"
				 }
			  ],
			  "ports":[  
				 {  
					"name":"insecure-connection",
					"port":80,
					"protocol":"UDP"
				 }
			  ]
		   }
		]
	 }`)

	var resource, overlay interface{}

	json.Unmarshal(resourceRaw, &resource)
	json.Unmarshal(overlayRaw, &overlay)

	patches, err := applyOverlay(resource, overlay, "/")
	assert.NilError(t, err)
	assert.Assert(t, patches != nil)

	patch := JoinPatches(patches)

	decoded, err := jsonpatch.DecodePatch(patch)
	assert.NilError(t, err)
	assert.Assert(t, decoded != nil)

	patched, err := decoded.Apply(resourceRaw)
	assert.NilError(t, err)
	assert.Assert(t, patched != nil)

	expectedResult := []byte(`{  
		"apiVersion":"v1",
		"kind":"Endpoints",
		"metadata":{  
		   "name":"test-endpoint",
		   "labels":{  
			  "label":"test"
		   }
		},
		"subsets":[  
		   {  
			  "addresses":[  
				 {  
					"ip":"192.168.10.171"
				 }
			  ],
			  "ports":[  
				 {  
					"name":"secure-connection",
					"port":443,
					"protocol":"TCP"
				 }
			  ]
		   },
		   {  
			  "addresses":[  
				 {  
					"ip":"192.168.10.172"
				 },
				 {  
					"ip":"192.168.10.173"
				 }
			  ],
			  "ports":[  
				 {  
					"name":"insecure-connection",
					"port":80,
					"protocol":"UDP"
				 }
			  ]
		   }
		]
	 }`)

	compareJSONAsMap(t, expectedResult, patched)
}

func TestApplyOverlay_TestInsertToArray(t *testing.T) {
	overlayRaw := []byte(`
	 {  
		"spec":{  
		   "template":{  
			  "spec":{  
				 "containers":[  
					{  
					   "name":"pi1",
					   "image":"vasylev.perl"
					}
				 ]
			  }
		   }
		}
	 }`)
	resourceRaw := []byte(`{  
		"apiVersion":"batch/v1",
		"kind":"Job",
		"metadata":{  
		   "name":"pi"
		},
		"spec":{  
		   "template":{  
			  "spec":{  
				 "containers":[  
					{  
					   "name":"piv0",
					   "image":"perl",
					   "command":[  
						  "perl"
					   ]
					},
					{  
					   "name":"pi",
					   "image":"perl",
					   "command":[  
						  "perl"
					   ]
					},
					{  
					   "name":"piv1",
					   "image":"perl",
					   "command":[  
						  "perl"
					   ]
					}
				 ],
				 "restartPolicy":"Never"
			  }
		   },
		   "backoffLimit":4
		}
	 }`)

	var resource, overlay interface{}

	json.Unmarshal(resourceRaw, &resource)
	json.Unmarshal(overlayRaw, &overlay)

	patches, err := applyOverlay(resource, overlay, "/")
	assert.NilError(t, err)
	assert.Assert(t, patches != nil)

	patch := JoinPatches(patches)

	decoded, err := jsonpatch.DecodePatch(patch)
	assert.NilError(t, err)
	assert.Assert(t, decoded != nil)

	patched, err := decoded.Apply(resourceRaw)
	assert.NilError(t, err)
	assert.Assert(t, patched != nil)
}

func TestApplyOverlay_ImagePullPolicy(t *testing.T) {
	overlayRaw := []byte(`{
		"spec": {
			"template": {
				"spec": {
					"containers": [
						{
							"(image)": "*:latest",
							"imagePullPolicy": "IfNotPresent",
							"ports": [
								{
									"containerPort": 8080
								}
							]
						}
					]
				}
			}
		}
	}`)
	resourceRaw := []byte(`{
		"apiVersion": "apps/v1",
		"kind": "Deployment",
		"metadata": {
			"name": "nginx-deployment",
			"labels": {
				"app": "nginx"
			}
		},
		"spec": {
			"replicas": 1,
			"selector": {
				"matchLabels": {
					"app": "nginx"
				}
			},
			"template": {
				"metadata": {
					"labels": {
						"app": "nginx"
					}
				},
				"spec": {
					"containers": [
						{
							"name": "nginx",
							"image": "nginx:latest",
							"ports": [
								{
									"containerPort": 80
								}
							]
						},
						{
							"name": "ghost",
							"image": "ghost:latest"
						}
					]
				}
			}
		}
	}`)

	var resource, overlay interface{}

	json.Unmarshal(resourceRaw, &resource)
	json.Unmarshal(overlayRaw, &overlay)

	patches, err := applyOverlay(resource, overlay, "/")
	assert.NilError(t, err)
	assert.Assert(t, len(patches) != 0)

	doc, err := ApplyPatches(resourceRaw, patches)
	assert.NilError(t, err)
	expectedResult := []byte(`{  
		"apiVersion":"apps/v1",
		"kind":"Deployment",
		"metadata":{  
		   "name":"nginx-deployment",
		   "labels":{  
			  "app":"nginx"
		   }
		},
		"spec":{  
		   "replicas":1,
		   "selector":{  
			  "matchLabels":{  
				 "app":"nginx"
			  }
		   },
		   "template":{  
			  "metadata":{  
				 "labels":{  
					"app":"nginx"
				 }
			  },
			  "spec":{  
				 "containers":[  
					{  
					   "image":"nginx:latest",
					   "imagePullPolicy":"IfNotPresent",
					   "name":"nginx",
					   "ports":[  
						  {  
							 "containerPort":80
						  },
						  {  
							 "containerPort":8080
						  }
					   ]
					},
					{  
					   "image":"ghost:latest",
					   "imagePullPolicy":"IfNotPresent",
					   "name":"ghost",
					   "ports":[  
						  {  
							 "containerPort":8080
						  }
					   ]
					}
				 ]
			  }
		   }
		}
	 }`)

	compareJSONAsMap(t, expectedResult, doc)
}

func TestApplyOverlay_AddingAnchor(t *testing.T) {
	overlayRaw := []byte(`{
		"metadata": {
			"name": "nginx-deployment",
			"labels": {
				"+(app)": "should-not-be-here",
				"+(key1)": "value1"
			}
		}
	}`)
	resourceRaw := []byte(`{
		"metadata": {
			"name": "nginx-deployment",
			"labels": {
				"app": "nginx"
			}
		}
	}`)

	var resource, overlay interface{}

	json.Unmarshal(resourceRaw, &resource)
	json.Unmarshal(overlayRaw, &overlay)

	patches, err := applyOverlay(resource, overlay, "/")
	assert.NilError(t, err)
	assert.Assert(t, len(patches) != 0)

	doc, err := ApplyPatches(resourceRaw, patches)
	assert.NilError(t, err)
	expectedResult := []byte(`{  
		"metadata":{  
		   "labels":{  
			  "app":"nginx",
			  "key1":"value1"
		   },
		   "name":"nginx-deployment"
		}
	 }`)

	compareJSONAsMap(t, expectedResult, doc)
}

func TestApplyOverlay_AddingAnchorInsideListElement(t *testing.T) {
	overlayRaw := []byte(`
	{
		"spec": {
			"template": {
				"spec": {
					"containers": [
						{
							"(image)": "*:latest",
							"+(imagePullPolicy)": "IfNotPresent"
						}
					]
				}
			}
		}
	}`)
	resourceRaw := []byte(`
	{  
		"apiVersion":"apps/v1",
		"kind":"Deployment",
		"metadata":{  
			"name":"nginx-deployment",
			"labels":{  
				"app":"nginx"
			}
		},
		"spec":{  
			"replicas":1,
			"selector":{  
				"matchLabels":{  
					"app":"nginx"
				}
			},
			"template":{  
				"metadata":{  
					"labels":{  
						"app":"nginx"
					}
				},
				"spec":{  
					"containers":[  
						{  
							"image":"nginx:latest"
						},
						{  
							"image":"ghost:latest",
							"imagePullPolicy":"Always"
						},
						{  
							"image":"debian:10"
						},
						{  
							"image":"ubuntu:18.04",
							"imagePullPolicy":"Always"
						}
					]
				}
			}
		}
	}`)

	var resource, overlay interface{}

	json.Unmarshal(resourceRaw, &resource)
	json.Unmarshal(overlayRaw, &overlay)

	patches, err := applyOverlay(resource, overlay, "/")
	assert.NilError(t, err)
	assert.Assert(t, len(patches) != 0)

	doc, err := ApplyPatches(resourceRaw, patches)
	assert.NilError(t, err)
	expectedResult := []byte(`
	{  
		"apiVersion":"apps/v1",
		"kind":"Deployment",
		"metadata":{  
			"name":"nginx-deployment",
			"labels":{  
				"app":"nginx"
			}
		},
		"spec":{  
			"replicas":1,
			"selector":{  
				"matchLabels":{  
					"app":"nginx"
				}
			},
			"template":{  
				"metadata":{  
					"labels":{  
						"app":"nginx"
					}
				},
				"spec":{  
					"containers":[  
						{  
							"image":"nginx:latest",
							"imagePullPolicy":"IfNotPresent"
						},
						{  
							"image":"ghost:latest",
							"imagePullPolicy":"Always"
						},
						{  
							"image":"debian:10"
						},
						{  
							"image":"ubuntu:18.04",
							"imagePullPolicy":"Always"
						}
					]
				}
			}
		}
	}`)
	compareJSONAsMap(t, expectedResult, doc)
}