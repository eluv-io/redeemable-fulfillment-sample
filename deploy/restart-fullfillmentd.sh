
podman ps -a

systemctl --user stop fulfillmentd && systemctl --user start fulfillmentd

sleep 2 
podman ps -a
