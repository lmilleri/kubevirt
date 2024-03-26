package network_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"kubevirt.io/kubevirt/pkg/network/downwardapi"

	virtv1 "kubevirt.io/api/core/v1"

	"kubevirt.io/kubevirt/pkg/network/deviceinfo"

	"kubevirt.io/kubevirt/pkg/virt-controller/network"
)

var _ = Describe("pod annotations", func() {

	Context("Generate pod network annotations", func() {
		const (
			deviceInfoPlugin    = "deviceinfo"
			nonDeviceInfoPlugin = "non_deviceinfo"
		)

		bindingPlugins := map[string]virtv1.InterfaceBindingPlugin{
			deviceInfoPlugin:    {DownwardAPI: virtv1.DeviceInfo},
			nonDeviceInfoPlugin: {},
		}

		networkStatus := `
[
{
 "name": "default/no-device-info",
 "interface": "pod6446d58d6df",
 "mac": "8a:37:d9:e7:0f:18",
 "dns": {}
},
{
 "name": "default/with-device-info",
 "interface": "pod2c26b46b68f",
 "dns": {},
 "device-info": {
   "type": "pci",
   "version": "1.0.0",
   "pci": {
     "pci-address": "0000:65:00.2"
   }
 }
},
{
 "name": "default/sriov",
 "interface": "pod778c553efa0",
 "dns": {},
 "device-info": {
   "type": "pci",
   "version": "1.0.0",
   "pci": {
     "pci-address": "0000:65:00.3"
   }
 }
}
]`

		It("should be empty when there are no networks", func() {
			networks := []virtv1.Network{}

			interfaces := []virtv1.Interface{}
			podAnnotationMap := network.GeneratePodAnnotations(networks, interfaces, networkStatus, bindingPlugins)
			Expect(podAnnotationMap).To(BeEmpty())
		})
		It("should be empty when there are no networks with binding plugin/SRIOV", func() {
			networks := []virtv1.Network{
				newMultusNetwork("boo", "default/no-device-info"),
			}

			interfaces := []virtv1.Interface{
				newInterface("boo"),
			}
			podAnnotationMap := network.GeneratePodAnnotations(networks, interfaces, networkStatus, bindingPlugins)
			Expect(podAnnotationMap).To(BeEmpty())
		})
		It("should be empty when there are networks with binding plugin but none with device-info", func() {
			networks := []virtv1.Network{
				newMultusNetwork("boo", "default/no-device-info"),
			}

			interfaces := []virtv1.Interface{
				newBindingPluginInterface("boo", nonDeviceInfoPlugin),
			}
			podAnnotationMap := network.GeneratePodAnnotations(networks, interfaces, networkStatus, bindingPlugins)
			Expect(podAnnotationMap).To(BeEmpty())
		})
		It("should have network-info entry when there is one non SRIOV interface with device info", func() {
			networks := []virtv1.Network{
				newMultusNetwork("foo", "default/with-device-info"),
			}

			interfaces := []virtv1.Interface{
				newBindingPluginInterface("foo", deviceInfoPlugin),
			}
			podAnnotationMap := network.GeneratePodAnnotations(networks, interfaces, networkStatus, bindingPlugins)
			Expect(podAnnotationMap).To(HaveLen(1))
			Expect(podAnnotationMap).To(HaveKeyWithValue(downwardapi.NetworkInfoAnnot, `{"interfaces":[{"network":"foo","deviceInfo":{"type":"pci","version":"1.0.0","pci":{"pci-address":"0000:65:00.2"}}}]}`))
		})
		It("should have both network-pci-map, network-info entries when there is SRIOV interface and binding plugin interface with device-info", func() {
			networks := []virtv1.Network{
				newMultusNetwork("foo", "default/with-device-info"),
				newMultusNetwork("doo", "default/sriov"),
			}

			interfaces := []virtv1.Interface{
				newBindingPluginInterface("foo", deviceInfoPlugin),
				newSRIOVInterface("doo"),
			}
			podAnnotationMap := network.GeneratePodAnnotations(networks, interfaces, networkStatus, bindingPlugins)
			Expect(podAnnotationMap).To(HaveLen(2))
			Expect(podAnnotationMap).To(HaveKeyWithValue(downwardapi.NetworkInfoAnnot, `{"interfaces":[{"network":"foo","deviceInfo":{"type":"pci","version":"1.0.0","pci":{"pci-address":"0000:65:00.2"}}}]}`))
			Expect(podAnnotationMap).To(HaveKeyWithValue(deviceinfo.NetworkPCIMapAnnot, `{"doo":"0000:65:00.3"}`))
		})
	})
})

func newMultusNetwork(name, networkName string) virtv1.Network {
	return virtv1.Network{
		Name: name,
		NetworkSource: virtv1.NetworkSource{
			Multus: &virtv1.MultusNetwork{
				NetworkName: networkName,
			},
		},
	}
}

func newInterface(name string) virtv1.Interface {
	return virtv1.Interface{
		Name: name,
	}
}

func newSRIOVInterface(name string) virtv1.Interface {
	return virtv1.Interface{
		Name:                   name,
		InterfaceBindingMethod: virtv1.InterfaceBindingMethod{SRIOV: &virtv1.InterfaceSRIOV{}},
	}
}

func newBindingPluginInterface(name, bindingPlugin string) virtv1.Interface {
	return virtv1.Interface{
		Name:    name,
		Binding: &virtv1.PluginBinding{Name: bindingPlugin},
	}
}
