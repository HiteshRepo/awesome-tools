## Issue: Helm Release Lock in Pulumi Deployments

### Symptom

During a Pulumi deployment involving a Helm release, the pipeline failed with:

```bash
kubernetes:helm.sh/v3:Release (control-plane)
error: another operation (install/upgrade/rollback) is in progress
```

### Root Cause

Helm stores release state in Kubernetes Secrets (or ConfigMaps) named like:

```bash
sh.helm.release.v1.<release-name>.v<revision>
```

If a deployment is interrupted (e.g., pipeline cancel, crash, network error), Helm may leave the most recent revision in a pending-upgrade or pending-install state.
- helm list may show no release because it considers the release incomplete.
- Pulumi retries and encounters the lock, failing with the above error.

### Resolution

Identify secrets related to the stuck release:

```bash
kubectl get secrets -n A | grep control-plane
```

Inspect the latest revision:

```bash
kubectl describe secret sh.helm.release.v1.control-plane.vX -n <namespace> | grep status:
```

If status shows pending-upgrade or pending-install, it’s blocking the release.


Delete only the latest stuck revision:

```bash
kubectl delete secret sh.helm.release.v1.control-plane.vX -n <namespace>
```

Retry the Pulumi deployment:

```bash
pulumi up
```

### Prevention

- Avoid canceling Pulumi jobs mid-deploy.
- Configure concurrency control in GitHub Actions so only one deployment per stack runs at a time.
- Optionally, create a cleanup script to detect and clear pending-* Helm secrets when needed.