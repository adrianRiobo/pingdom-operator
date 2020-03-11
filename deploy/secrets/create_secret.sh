username=$(echo -n $1 | base64 -w0)
password=$(echo -n $2 | base64 -w0)
apikey=$(echo -n $3 | base64 -w0)

#Direct apply
# | kubectl apply -f -
cat <<EOF > pingdomsecret.yaml
---
apiVersion: v1
kind: Secret
metadata:
  name: pingdom-credentials
type: Opaque
data:
  username: ${username}
  password: ${password}
  apikey:   ${apikey}
EOF
