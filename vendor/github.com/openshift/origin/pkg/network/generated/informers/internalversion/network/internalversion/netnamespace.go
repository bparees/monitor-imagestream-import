// This file was automatically generated by informer-gen

package internalversion

import (
	network "github.com/openshift/origin/pkg/network/apis/network"
	internalinterfaces "github.com/openshift/origin/pkg/network/generated/informers/internalversion/internalinterfaces"
	internalclientset "github.com/openshift/origin/pkg/network/generated/internalclientset"
	internalversion "github.com/openshift/origin/pkg/network/generated/listers/network/internalversion"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	runtime "k8s.io/apimachinery/pkg/runtime"
	watch "k8s.io/apimachinery/pkg/watch"
	cache "k8s.io/client-go/tools/cache"
	time "time"
)

// NetNamespaceInformer provides access to a shared informer and lister for
// NetNamespaces.
type NetNamespaceInformer interface {
	Informer() cache.SharedIndexInformer
	Lister() internalversion.NetNamespaceLister
}

type netNamespaceInformer struct {
	factory          internalinterfaces.SharedInformerFactory
	tweakListOptions internalinterfaces.TweakListOptionsFunc
}

// NewNetNamespaceInformer constructs a new informer for NetNamespace type.
// Always prefer using an informer factory to get a shared informer instead of getting an independent
// one. This reduces memory footprint and number of connections to the server.
func NewNetNamespaceInformer(client internalclientset.Interface, resyncPeriod time.Duration, indexers cache.Indexers) cache.SharedIndexInformer {
	return NewFilteredNetNamespaceInformer(client, resyncPeriod, indexers, nil)
}

// NewFilteredNetNamespaceInformer constructs a new informer for NetNamespace type.
// Always prefer using an informer factory to get a shared informer instead of getting an independent
// one. This reduces memory footprint and number of connections to the server.
func NewFilteredNetNamespaceInformer(client internalclientset.Interface, resyncPeriod time.Duration, indexers cache.Indexers, tweakListOptions internalinterfaces.TweakListOptionsFunc) cache.SharedIndexInformer {
	return cache.NewSharedIndexInformer(
		&cache.ListWatch{
			ListFunc: func(options v1.ListOptions) (runtime.Object, error) {
				if tweakListOptions != nil {
					tweakListOptions(&options)
				}
				return client.Network().NetNamespaces().List(options)
			},
			WatchFunc: func(options v1.ListOptions) (watch.Interface, error) {
				if tweakListOptions != nil {
					tweakListOptions(&options)
				}
				return client.Network().NetNamespaces().Watch(options)
			},
		},
		&network.NetNamespace{},
		resyncPeriod,
		indexers,
	)
}

func (f *netNamespaceInformer) defaultInformer(client internalclientset.Interface, resyncPeriod time.Duration) cache.SharedIndexInformer {
	return NewFilteredNetNamespaceInformer(client, resyncPeriod, cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc}, f.tweakListOptions)
}

func (f *netNamespaceInformer) Informer() cache.SharedIndexInformer {
	return f.factory.InformerFor(&network.NetNamespace{}, f.defaultInformer)
}

func (f *netNamespaceInformer) Lister() internalversion.NetNamespaceLister {
	return internalversion.NewNetNamespaceLister(f.Informer().GetIndexer())
}
