package network

import (
	networkv1 "github.com/k8snetworkplumbingwg/network-attachment-definition-client/pkg/apis/k8s.cni.cncf.io/v1"

	virtv1 "kubevirt.io/api/core/v1"
	"kubevirt.io/client-go/log"

	"kubevirt.io/kubevirt/pkg/network/deviceinfo"
	"kubevirt.io/kubevirt/pkg/network/downwardapi"
	"kubevirt.io/kubevirt/pkg/network/vmispec"
)

func GeneratePodAnnotations(networks []virtv1.Network, interfaces []virtv1.Interface, multusStatusAnnotation string, bindingPlugins map[string]virtv1.InterfaceBindingPlugin) map[string]string {
	newAnnotations := map[string]string{}
	if vmispec.SRIOVInterfaceExist(interfaces) {
		networkPCIMapAnnotationValue := deviceinfo.CreateNetworkPCIAnnotationValue(
			networks, interfaces, multusStatusAnnotation,
		)
		newAnnotations[deviceinfo.NetworkPCIMapAnnot] = networkPCIMapAnnotationValue
	}
	if vmispec.BindingPluginNetworkWithDeviceInfoExist(interfaces, bindingPlugins) {
		networkDeviceInfoMap, err := deviceinfo.MapBindingPluginNetworkNameToDeviceInfo(networks, interfaces, multusStatusAnnotation, bindingPlugins)
		if err != nil {
			log.Log.Warningf("failed to create network-device-info-map: %v", err)
			networkDeviceInfoMap = map[string]*networkv1.DeviceInfo{}
		}
		networkDeviceInfoAnnotation := downwardapi.CreateNetworkInfoAnnotationValue(networkDeviceInfoMap)
		newAnnotations[downwardapi.NetworkInfoAnnot] = networkDeviceInfoAnnotation
	}

	return newAnnotations
}
