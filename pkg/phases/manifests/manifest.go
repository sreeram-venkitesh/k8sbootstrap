package manifests

import (
	"fmt"
	"os"

	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/serializer/json"
	"k8s.io/client-go/kubernetes/scheme"
)

var hostPathDirectoryOrCreate = v1.HostPathDirectoryOrCreate
var hostPathFileOrCreate = v1.HostPathFileOrCreate
var hostname, _ = os.Hostname()

func SetupStaticPodManifests(advertiseAddress string) error {
	err := os.MkdirAll("/etc/kubernetes/manifests", 0755)
	if err != nil {
		return fmt.Errorf("failed to create manifests directory: %w", err)
	}

	err = SetupEtcdStaticPodManifest(advertiseAddress)
	if err != nil {
		return fmt.Errorf("failed to etcd pod manifest: %w", err)
	}

	err = SetupApiserverStaticPodManifest(advertiseAddress)
	if err != nil {
		return fmt.Errorf("failed to create apiserver pod manifest: %w", err)
	}

	err = SetupControllerManagerStaticPodManifest()
	if err != nil {
		return fmt.Errorf("failed to create controller manager pod manifest: %w", err)
	}

	err = SetupSchedulerStaticPodManifest()
	if err != nil {
		return fmt.Errorf("failed to create scheduler pod manifest: %w", err)
	}

	return nil
}

func SetupApiserverStaticPodManifest(advertiseAddress string) error {
	apiserverPod := &corev1.Pod{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Pod",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "kube-apiserver",
			Namespace: "kube-system",
			Labels: map[string]string{
				"component": "kube-apiserver",
				"tier":      "control-plane",
			},
		},
		Spec: corev1.PodSpec{
			HostNetwork:       true,
			PriorityClassName: "system-node-critical",
			RestartPolicy:     corev1.RestartPolicyAlways,
			Containers: []corev1.Container{
				{
					Name:  "kube-apiserver",
					Image: "registry.k8s.io/kube-apiserver:v1.35.0",
					Command: []string{
						"kube-apiserver",
						fmt.Sprintf("--advertise-address=%s", advertiseAddress),
						"--allow-privileged=true",
						"--bind-address=0.0.0.0",
						"--authorization-mode=Node,RBAC",
						"--client-ca-file=/etc/kubernetes/pki/ca.crt",
						"--enable-admission-plugins=NodeRestriction",
						"--enable-bootstrap-token-auth=true",
						"--etcd-cafile=/etc/kubernetes/pki/ca.crt",
						"--etcd-certfile=/etc/kubernetes/pki/apiserver-etcd-client.crt",
						"--etcd-keyfile=/etc/kubernetes/pki/apiserver-etcd-client.key",
						"--etcd-servers=https://127.0.0.1:2379",
						"--kubelet-client-certificate=/etc/kubernetes/pki/apiserver-kubelet-client.crt",
						"--kubelet-client-key=/etc/kubernetes/pki/apiserver-kubelet-client.key",
						"--kubelet-preferred-address-types=InternalIP,ExternalIP,Hostname",
						// "--proxy-client-cert-file=/etc/kubernetes/pki/front-proxy-client.crt",
						// "--proxy-client-key-file=/etc/kubernetes/pki/front-proxy-client.key",
						// "--requestheader-allowed-names=front-proxy-client",
						// "--requestheader-client-ca-file=/etc/kubernetes/pki/front-proxy-ca.crt",
						"--requestheader-extra-headers-prefix=X-Remote-Extra-",
						"--requestheader-group-headers=X-Remote-Group",
						"--requestheader-username-headers=X-Remote-User",
						"--runtime-config=",
						"--secure-port=6443",
						"--service-account-issuer=https://kubernetes.default.svc.cluster.local",
						"--service-account-key-file=/etc/kubernetes/pki/sa.pub",
						"--service-account-signing-key-file=/etc/kubernetes/pki/sa.key",
						"--service-cluster-ip-range=10.96.0.0/16",
						"--tls-cert-file=/etc/kubernetes/pki/apiserver.crt",
						"--tls-private-key-file=/etc/kubernetes/pki/apiserver.key",
					},
					VolumeMounts: []v1.VolumeMount{
						{
							Name:      "k8s-certs",
							MountPath: "/etc/kubernetes/pki/",
							ReadOnly:  true,
						},
					},
				},
			},
			Volumes: []corev1.Volume{
				{
					Name: "k8s-certs",
					VolumeSource: v1.VolumeSource{
						HostPath: &v1.HostPathVolumeSource{
							Path: "/etc/kubernetes/pki/",
							Type: &hostPathDirectoryOrCreate,
						},
					},
				},
			},
		},
	}

	err := writePodManifest(apiserverPod, "/etc/kubernetes/manifests/kube-apiserver.yaml")
	if err != nil {
		return fmt.Errorf("failed to write apiserver manifest: %w", err)
	}

	fmt.Printf("[manifest] kube-apiserver static pod manifest succesfully generated\n")
	return nil
}

func SetupControllerManagerStaticPodManifest() error {
	controllerManagerPod := &corev1.Pod{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Pod",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "kube-controller-manager",
			Namespace: "kube-system",
			Labels: map[string]string{
				"component": "kube-controller-manager",
				"tier":      "control-plane",
			},
		},
		Spec: corev1.PodSpec{
			HostNetwork:       true,
			PriorityClassName: "system-node-critical",
			RestartPolicy:     corev1.RestartPolicyAlways,
			Containers: []corev1.Container{
				{
					Name:  "kube-conrtoller-manager",
					Image: "registry.k8s.io/kube-controller-manager:v1.35.0",
					Command: []string{
						"kube-controller-manager",
						"--allocate-node-cidrs=true",
						"--authentication-kubeconfig=/etc/kubernetes/controller-manager.conf",
						"--authorization-kubeconfig=/etc/kubernetes/controller-manager.conf",
						"--bind-address=127.0.0.1",
						"--client-ca-file=/etc/kubernetes/pki/ca.crt",
						"--cluster-cidr=10.244.0.0/16",
						"--cluster-name=kubernetes",
						"--cluster-signing-cert-file=/etc/kubernetes/pki/ca.crt",
						"--cluster-signing-key-file=/etc/kubernetes/pki/ca.key",
						"--controllers=*,bootstrapsigner,tokencleaner",
						"--enable-hostpath-provisioner=true",
						"--kubeconfig=/etc/kubernetes/controller-manager.conf",
						"--leader-elect=true",
						// "--requestheader-client-ca-file=/etc/kubernetes/pki/front-proxy-ca.crt",
						"--root-ca-file=/etc/kubernetes/pki/ca.crt",
						"--service-account-private-key-file=/etc/kubernetes/pki/sa.key",
						"--service-cluster-ip-range=10.96.0.0/16",
						"--use-service-account-credentials=true",
					},
					VolumeMounts: []v1.VolumeMount{
						{
							Name:      "kubeconfig",
							MountPath: "/etc/kubernetes/controller-manager.conf",
							ReadOnly:  true,
						},
						{
							Name:      "k8s-certs",
							MountPath: "/etc/kubernetes/pki",
							ReadOnly:  true,
						},
					},
				},
			},
			Volumes: []corev1.Volume{
				{
					Name: "kubeconfig",
					VolumeSource: v1.VolumeSource{
						HostPath: &v1.HostPathVolumeSource{
							Path: "/etc/kubernetes/controller-manager.conf",
							Type: &hostPathFileOrCreate,
						},
					},
				},
				{
					Name: "k8s-certs",
					VolumeSource: v1.VolumeSource{
						HostPath: &v1.HostPathVolumeSource{
							Path: "/etc/kubernetes/pki/",
							Type: &hostPathDirectoryOrCreate,
						},
					},
				},
			},
		},
	}

	err := writePodManifest(controllerManagerPod, "/etc/kubernetes/manifests/kube-controller-manager.yaml")
	if err != nil {
		return fmt.Errorf("failed to write controller manager manifest: %w", err)
	}

	fmt.Printf("[manifest] kube-controller-manager static pod manifest succesfully generated\n")
	return nil
}

func SetupSchedulerStaticPodManifest() error {
	schedulerPod := &corev1.Pod{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Pod",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "kube-scheduler",
			Namespace: "kube-system",
			Labels: map[string]string{
				"component": "kube-scheduler",
				"tier":      "control-plane",
			},
		},
		Spec: corev1.PodSpec{
			HostNetwork:       true,
			PriorityClassName: "system-node-critical",
			RestartPolicy:     corev1.RestartPolicyAlways,
			Containers: []corev1.Container{
				{
					Name:  "kube-apiserver",
					Image: "registry.k8s.io/kube-scheduler:v1.35.0",
					Command: []string{
						"kube-scheduler",
						"--authentication-kubeconfig=/etc/kubernetes/scheduler.conf",
						"--authorization-kubeconfig=/etc/kubernetes/scheduler.conf",
						"--bind-address=127.0.0.1",
						"--kubeconfig=/etc/kubernetes/scheduler.conf",
						"--leader-elect=true",
					},
					VolumeMounts: []v1.VolumeMount{
						{
							Name:      "kubeconfig",
							MountPath: "/etc/kubernetes/scheduler.conf",
							ReadOnly:  true,
						},
					},
				},
			},
			Volumes: []corev1.Volume{
				{
					Name: "kubeconfig",
					VolumeSource: v1.VolumeSource{
						HostPath: &v1.HostPathVolumeSource{
							Path: "/etc/kubernetes/scheduler.conf",
							Type: &hostPathFileOrCreate,
						},
					},
				},
			},
		},
	}

	err := writePodManifest(schedulerPod, "/etc/kubernetes/manifests/kube-scheduler.yaml")
	if err != nil {
		return fmt.Errorf("failed to write scheduler manifest: %w", err)
	}

	fmt.Printf("[manifest] kube-scheduler static pod manifest succesfully generated\n")
	return nil
}

func SetupEtcdStaticPodManifest(advertiseAddress string) error {
	etcdPod := &corev1.Pod{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Pod",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "etcd",
			Namespace: "kube-system",
			Labels: map[string]string{
				"component": "etcd",
				"tier":      "control-plane",
			},
		},
		Spec: corev1.PodSpec{
			HostNetwork:       true,
			PriorityClassName: "system-node-critical",
			RestartPolicy:     corev1.RestartPolicyAlways,
			Containers: []corev1.Container{
				{
					Name:  "etcd",
					Image: "registry.k8s.io/etcd:3.6.6-0",
					Command: []string{
						"etcd",
						fmt.Sprintf("--advertise-client-urls=https://%s:2379", advertiseAddress),
						"--cert-file=/etc/kubernetes/pki/etcd/server.crt",
						"--client-cert-auth=true",
						"--data-dir=/var/lib/etcd",
						"--experimental-initial-corrupt-check=true",
						"--experimental-watch-progress-notify-interval=5s",
						// fmt.Sprintf("--initial-advertise-peer-urls=https://%s:2380", advertiseAddress),
						// fmt.Sprintf("--initial-cluster=%s=https://%s:2380", hostname, advertiseAddress),
						"--key-file=/etc/kubernetes/pki/etcd/server.key",
						fmt.Sprintf("--listen-client-urls=https://127.0.0.1:2379,https://%s:2379", advertiseAddress),
						"--listen-metrics-urls=http://127.0.0.1:2381",
						// fmt.Sprintf("--listen-peer-urls=https://%s:2380", advertiseAddress),
						fmt.Sprintf("--name=%s", hostname),
						// "--peer-cert-file=/etc/kubernetes/pki/etcd/peer.crt",
						// "--peer-client-cert-auth=false",
						// "--peer-key-file=/etc/kubernetes/pki/etcd/peer.key",
						// "--peer-trusted-ca-file=/etc/kubernetes/pki/ca.crt",
						"--snapshot-count=10000",
						"--trusted-ca-file=/etc/kubernetes/pki/ca.crt",
					},
					VolumeMounts: []v1.VolumeMount{
						{
							Name:      "etcd-certs",
							MountPath: "/etc/kubernetes/pki/",
						},
						{
							Name:      "etcd-data",
							MountPath: "/var/lib/etcd",
						},
					},
				},
			},
			Volumes: []corev1.Volume{
				{
					Name: "etcd-certs",
					VolumeSource: v1.VolumeSource{
						HostPath: &v1.HostPathVolumeSource{
							Path: "/etc/kubernetes/pki/",
							Type: &hostPathDirectoryOrCreate,
						},
					},
				},
				{
					Name: "etcd-data",
					VolumeSource: v1.VolumeSource{
						HostPath: &v1.HostPathVolumeSource{
							Path: "/var/lib/etcd",
							Type: &hostPathDirectoryOrCreate,
						},
					},
				},
			},
		},
	}

	err := writePodManifest(etcdPod, "/etc/kubernetes/manifests/etcd.yaml")
	if err != nil {
		return fmt.Errorf("failed to write etcd manifest: %w", err)
	}

	fmt.Printf("[manifest] etcd static pod manifest succesfully generated\n")
	return nil
}

func writePodManifest(pod *corev1.Pod, filename string) error {
	serializer := json.NewSerializerWithOptions(
		json.DefaultMetaFactory,
		scheme.Scheme,
		scheme.Scheme,
		json.SerializerOptions{Yaml: true, Pretty: true, Strict: true},
	)

	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	err = serializer.Encode(pod, file)
	if err != nil {
		return err
	}

	return nil
}
