package v1

import (
	v1 "github.com/jenkins-x/jx/pkg/apis/jenkins.io/v1"
	scheme "github.com/jenkins-x/jx/pkg/client/clientset/versioned/scheme"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	rest "k8s.io/client-go/rest"
)

// EnvironmentsGetter has a method to return a EnvironmentInterface.
// A group's client should implement this interface.
type EnvironmentsGetter interface {
	Environments(namespace string) EnvironmentInterface
}

// EnvironmentInterface has methods to work with Environment resources.
type EnvironmentInterface interface {
	Create(*v1.Environment) (*v1.Environment, error)
	Update(*v1.Environment) (*v1.Environment, error)
	Delete(name string, options *meta_v1.DeleteOptions) error
	DeleteCollection(options *meta_v1.DeleteOptions, listOptions meta_v1.ListOptions) error
	Get(name string, options meta_v1.GetOptions) (*v1.Environment, error)
	List(opts meta_v1.ListOptions) (*v1.EnvironmentList, error)
	Watch(opts meta_v1.ListOptions) (watch.Interface, error)
	Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1.Environment, err error)
	EnvironmentExpansion
}

// environments implements EnvironmentInterface
type environments struct {
	client rest.Interface
	ns     string
}

// newEnvironments returns a Environments
func newEnvironments(c *JenkinsV1Client, namespace string) *environments {
	return &environments{
		client: c.RESTClient(),
		ns:     namespace,
	}
}

// Get takes name of the environment, and returns the corresponding environment object, and an error if there is any.
func (c *environments) Get(name string, options meta_v1.GetOptions) (result *v1.Environment, err error) {
	result = &v1.Environment{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("environments").
		Name(name).
		VersionedParams(&options, scheme.ParameterCodec).
		Do().
		Into(result)
	return
}

// List takes label and field selectors, and returns the list of Environments that match those selectors.
func (c *environments) List(opts meta_v1.ListOptions) (result *v1.EnvironmentList, err error) {
	result = &v1.EnvironmentList{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("environments").
		VersionedParams(&opts, scheme.ParameterCodec).
		Do().
		Into(result)
	return
}

// Watch returns a watch.Interface that watches the requested environments.
func (c *environments) Watch(opts meta_v1.ListOptions) (watch.Interface, error) {
	opts.Watch = true
	return c.client.Get().
		Namespace(c.ns).
		Resource("environments").
		VersionedParams(&opts, scheme.ParameterCodec).
		Watch()
}

// Create takes the representation of a environment and creates it.  Returns the server's representation of the environment, and an error, if there is any.
func (c *environments) Create(environment *v1.Environment) (result *v1.Environment, err error) {
	result = &v1.Environment{}
	err = c.client.Post().
		Namespace(c.ns).
		Resource("environments").
		Body(environment).
		Do().
		Into(result)
	return
}

// Update takes the representation of a environment and updates it. Returns the server's representation of the environment, and an error, if there is any.
func (c *environments) Update(environment *v1.Environment) (result *v1.Environment, err error) {
	result = &v1.Environment{}
	err = c.client.Put().
		Namespace(c.ns).
		Resource("environments").
		Name(environment.Name).
		Body(environment).
		Do().
		Into(result)
	return
}

// Delete takes name of the environment and deletes it. Returns an error if one occurs.
func (c *environments) Delete(name string, options *meta_v1.DeleteOptions) error {
	return c.client.Delete().
		Namespace(c.ns).
		Resource("environments").
		Name(name).
		Body(options).
		Do().
		Error()
}

// DeleteCollection deletes a collection of objects.
func (c *environments) DeleteCollection(options *meta_v1.DeleteOptions, listOptions meta_v1.ListOptions) error {
	return c.client.Delete().
		Namespace(c.ns).
		Resource("environments").
		VersionedParams(&listOptions, scheme.ParameterCodec).
		Body(options).
		Do().
		Error()
}

// Patch applies the patch and returns the patched environment.
func (c *environments) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1.Environment, err error) {
	result = &v1.Environment{}
	err = c.client.Patch(pt).
		Namespace(c.ns).
		Resource("environments").
		SubResource(subresources...).
		Name(name).
		Body(data).
		Do().
		Into(result)
	return
}
