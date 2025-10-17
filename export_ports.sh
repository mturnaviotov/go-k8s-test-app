namespace=todo
export NAME_BACK=$(kubectl get pods --namespace todo -l "app=todo" -o jsonpath="{.items[0].metadata.name}")
export NAME_FRONT=$(kubectl get pods --namespace todo -l "app=frontend" -o jsonpath="{.items[0].metadata.name}")
kubectl port-forward --namespace $namespace $NAME_BACK 8081:8080 &
kubectl port-forward --namespace $namespace $NAME_FRONT 8080:80 &
echo "Backend port forwarded to http://localhost:8081"
echo "Frontend port forwarded to http://localhost:8080"
echo "Press [ENTER] to stop port forwarding..."
read
pkill -f "kubectl port-forward --namespace $namespace $NAME_BACK 8081:8080"
pkill -f "kubectl port-forward --namespace $namespace $NAME_FRONT 8080:80"
echo "Port forwarding stopped." 

