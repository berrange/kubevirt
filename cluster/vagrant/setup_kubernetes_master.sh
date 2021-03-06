#/bin/bash -xe

export KUBERNETES_MASTER=true
bash ./setup_kubernetes_common.sh

# Cockpit with kubernetes plugin
yum install -y cockpit cockpit-kubernetes
systemctl enable cockpit.socket && systemctl start cockpit.socket

# Create the master
kubeadm init --pod-network-cidr=10.244.0.0/16 --token abcdef.1234567890123456 --apiserver-advertise-address=$ADVERTISED_MASTER_IP

# Tell kubectl which config to use
export KUBECONFIG=/etc/kubernetes/admin.conf

set +e

kubectl version
while [ $? -ne 0 ]; do
  sleep 60
  echo 'Waiting for Kubernetes cluster to become functional...'
  kubectl version
done

set -e

# Work around https://github.com/kubernetes/kubernetes/issues/34101
# Weave otherwise the network provider does not work
kubectl -n kube-system get ds -l 'k8s-app=kube-proxy' -o json \
        | jq '.items[0].spec.template.spec.containers[0].command |= .+ ["--proxy-mode=userspace"]' \
        |   kubectl apply -f - && kubectl -n kube-system delete pods -l 'k8s-app=kube-proxy'

if [ "$NETWORK_PROVIDER" == "weave" ]; then 
  kubectl apply -f https://github.com/weaveworks/weave/releases/download/v1.9.4/weave-daemonset-k8s-1.6.yaml
else
  kubectl create -f kube-$NETWORK_PROVIDER.yaml
fi

# Allow scheduling pods on master
# Ignore retval because it might not be dedicated already
kubectl taint nodes master node-role.kubernetes.io/master:NoSchedule- || :

# TODO better scope the permissions, for now allow the default account everything
kubectl create clusterrolebinding add-on-cluster-admin --clusterrole=cluster-admin --serviceaccount=kube-system:default
kubectl create clusterrolebinding add-on-default-admin --clusterrole=cluster-admin --serviceaccount=default:default

mkdir -p /exports/share1

chmod 0755 /exports/share1
chown 36:36 /exports/share1

echo "/exports/share1  *(rw,anonuid=36,anongid=36,all_squash,sync,no_subtree_check)" > /etc/exports

systemctl enable nfs-server && systemctl start nfs-server
