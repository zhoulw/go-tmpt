package main

import (
	"context"
	"flag"
	"fmt"
	"go-tmpt/action"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
	"path/filepath"
)

func main() {
	var kubeconfig *string
	var actionFlag *string

	// 试图取到当前账号的家目录
	if home := homedir.HomeDir(); home != "" {
		// 如果能取到，就把家目录下的.kube/config作为默认配置文件
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		// 如果取不到，就没有默认配置文件，必须通过kubeconfig参数来指定
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}

	actionFlag = flag.String("action", "list-pod", "指定实际操作功能")

	flag.Parse()

	fmt.Println("解析命令完毕，开始加载配置文件")

	// 加载配置文件
	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		panic(err.Error())
	}

	// 用clientset类来执行后续的查询操作
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	fmt.Printf("加载配置文件完毕，即将执行业务 [%v]\n", *actionFlag)

	var actionInterface action.Action

	// 注意，如果有新的功能类实现，就在这里添加对应的处理
	switch *actionFlag {
	case "list-pod":
		listPod := action.ListPod{}
		actionInterface = &listPod
	case "conflict":
		conflict := action.Confilct{}
		actionInterface = &conflict
	}

	err = actionInterface.DoAction(clientset)
	if err != nil {
		fmt.Printf("err: %v\n", err)
	} else {
		fmt.Println("执行完成")
	}
}

func main1() {
	var kubeconfig *string
	if home := homedir.HomeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}
	flag.Parse()
	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		panic(err.Error())
	}

	// create the clientset
	clientset, err := kubernetes.NewForConfig(config)

	// dynamic client
	dynamicCli, err := dynamic.NewForConfig(config)

	if err != nil {
		panic(err.Error())
	}

	gvr := schema.GroupVersionResource{
		Version:  "v1",
		Resource: "pods",
	}
	//``````````````````````````````````````````````````````````````````````````````````````````````````````````````````````
	// 使用dynamicClient的查询列表方法，查询指定namespace下的所有pod，
	// 注意此方法返回的数据结构类型是UnstructuredList
	unstructObj, err := dynamicCli.Resource(gvr).Namespace("kube-system").List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		panic(err.Error())
	}

	podList := &v1.PodList{}

	err = runtime.DefaultUnstructuredConverter.FromUnstructured(unstructObj.UnstructuredContent(), podList)
	if err != nil {
		panic(err.Error())
	}

	fmt.Printf("namespace\t status\t\t name\n")
	for _, data := range podList.Items {
		fmt.Printf("%v\t %v\t %v\n", data.Namespace, data.Status.Phase, data.Name)
	}

	//fmt.Println(clientset.ServerVersion())

	//namespaceList, _ := clientset.CoreV1().Namespaces().List(context.Background(), metav1.ListOptions{})
	//for _, namespace := range namespaceList.Items {
	//	fmt.Println(namespace)
	//	break
	//}

	//````````````````````````````````````````````````````````````````````````````````````````````````````````````````
	//client Set 方式获取
	podLists, _ := clientset.CoreV1().Pods("ns1").List(context.Background(), metav1.ListOptions{})

	for _, pod := range podLists.Items {
		fmt.Println(pod)
		break
	}

	//````````````````````````````````````````````````````````````````````````````````````````````````````````````````
	//discoverySClient聚焦于资源而不是资源对象，例如查看当前对象有哪些group、version、resource
	discoveryClient, err := discovery.NewDiscoveryClientForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	// 获取所有分组和资源数据
	APIGroup, APIResourceListSlice, err := discoveryClient.ServerGroupsAndResources()

	if err != nil {
		panic(err.Error())
	}

	// 先看Group信息
	fmt.Printf("APIGroup :\n\n %v\n\n\n\n", APIGroup)

	// APIResourceListSlice是个切片，里面的每个元素代表一个GroupVersion及其资源
	for _, singleAPIResourceList := range APIResourceListSlice {

		// GroupVersion是个字符串，例如"apps/v1"
		groupVerionStr := singleAPIResourceList.GroupVersion

		// ParseGroupVersion方法将字符串转成数据结构
		gv, err := schema.ParseGroupVersion(groupVerionStr)

		if err != nil {
			panic(err.Error())
		}

		fmt.Println("*****************************************************************")
		fmt.Printf("GV string [%v]\nGV struct [%#v]\nresources :\n\n", groupVerionStr, gv)

		// APIResources字段是个切片，里面是当前GroupVersion下的所有资源
		for _, singleAPIResource := range singleAPIResourceList.APIResources {
			fmt.Printf("%v\n", singleAPIResource.Name)
		}

	}
}

type dynamicClient struct {
	client *rest.RESTClient
}
