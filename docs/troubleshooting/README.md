# KubeChat Troubleshooting Guide

This guide helps you resolve common issues with KubeChat's container-first development environment.

## 🚨 Quick Diagnostics

### Run These Commands First

```bash
# 1. Validate environment
./infrastructure/scripts/validate-env.sh

# 2. Check deployment status
make dev-status

# 3. View application logs
make dev-logs

# 4. Check system health
./infrastructure/scripts/health-check.sh
```

## 🔧 Common Issues

### Environment Setup Issues

#### Issue: Environment Validation Fails

**Symptoms:**
- `validate-env.sh` shows red ❌ marks
- Missing required tools or versions

**Solutions:**

```bash
# For macOS with Homebrew
brew install --cask rancher-desktop
brew install helm kubectl go node pnpm

# For Ubuntu/Debian
sudo apt update
sudo apt install docker.io kubectl golang-go nodejs npm

# Install PNPM
curl -fsSL https://get.pnpm.io/install.sh | sh

# Install Helm
curl https://raw.githubusercontent.com/helm/helm/main/scripts/get-helm-3 | bash
```

#### Issue: Rancher Desktop Not Working

**Symptoms:**
- `kubectl cluster-info` fails
- Docker commands not working

**Solutions:**

1. **Start Rancher Desktop:**
   - Open Rancher Desktop application
   - Enable Kubernetes in settings
   - Wait for initialization to complete

2. **Reset Rancher Desktop:**
   ```bash
   # Stop Rancher Desktop
   # Go to Settings > Reset > Reset Kubernetes
   # Restart Rancher Desktop
   ```

3. **Check Kubernetes Context:**
   ```bash
   kubectl config current-context
   # Should show: rancher-desktop
   
   kubectl config use-context rancher-desktop
   ```

### Deployment Issues

#### Issue: `make dev-deploy` Fails

**Symptoms:**
- Helm installation errors
- Pods stuck in pending state
- Container image pull failures

**Solutions:**

1. **Clean and Retry:**
   ```bash
   make dev-clean
   make dev-build
   make dev-deploy
   ```

2. **Check Kubernetes Resources:**
   ```bash
   kubectl get nodes
   kubectl get storageclass
   kubectl describe pod <failing-pod> -n kubechat
   ```

3. **Fix Storage Issues:**
   ```bash
   # Ensure local-path storage class exists
   kubectl get storageclass local-path
   
   # If missing, Rancher Desktop may need reset
   ```

#### Issue: Pods Stuck in Pending

**Symptoms:**
- Pods show "Pending" status for extended time
- Events show scheduling errors

**Solutions:**

1. **Check Node Resources:**
   ```bash
   kubectl describe nodes
   kubectl top nodes  # If metrics available
   ```

2. **Check Pod Events:**
   ```bash
   kubectl describe pod <pod-name> -n kubechat
   kubectl get events -n kubechat --sort-by='.firstTimestamp'
   ```

3. **Resource Constraints:**
   ```bash
   # Reduce resource requests in values-dev.yaml
   # Edit infrastructure/helm/kubechat/values-dev.yaml
   # Lower memory/CPU requests
   ```

#### Issue: Container Image Pull Failures

**Symptoms:**
- "ImagePullBackOff" or "ErrImagePull" status
- Cannot pull container images

**Solutions:**

1. **Rebuild Images:**
   ```bash
   make dev-build
   docker images | grep kubechat
   ```

2. **Check Docker Daemon:**
   ```bash
   docker info
   docker system prune -f  # Clean up space
   ```

3. **Network Issues:**
   ```bash
   # Test connectivity
   curl -I https://hub.docker.com
   
   # If behind corporate firewall, configure Docker proxy
   ```

### Database Issues

#### Issue: Database Connection Failures

**Symptoms:**
- API cannot connect to PostgreSQL
- Database-related errors in logs

**Solutions:**

1. **Check Database Pod:**
   ```bash
   kubectl get pods -n kubechat -l app.kubernetes.io/name=postgresql
   kubectl logs -n kubechat <postgres-pod-name>
   ```

2. **Test Database Connectivity:**
   ```bash
   make dev-db-connect
   # Should open psql prompt
   ```

3. **Reset Database:**
   ```bash
   make dev-db-reset
   # WARNING: This deletes all data
   ```

#### Issue: Database Migrations Fail

**Symptoms:**
- Migration script errors
- Schema version conflicts

**Solutions:**

```bash
# Check migration status
./infrastructure/scripts/database/migrate.sh status

# Reset and re-run migrations
./infrastructure/scripts/database/migrate.sh reset
./infrastructure/scripts/database/migrate.sh full
```

### Application Issues

#### Issue: Frontend Not Loading

**Symptoms:**
- Blank page at http://localhost:30001
- Console errors in browser

**Solutions:**

1. **Check Web Service:**
   ```bash
   kubectl get svc kubechat-dev-web -n kubechat
   kubectl get pods -n kubechat -l app.kubernetes.io/component=web
   ```

2. **Check Logs:**
   ```bash
   make dev-logs-web
   ```

3. **Rebuild Frontend:**
   ```bash
   make dev-rebuild-web
   kubectl rollout restart deployment/kubechat-dev-web -n kubechat
   ```

#### Issue: API Endpoints Not Responding

**Symptoms:**
- 404 or 500 errors from API
- API health check fails

**Solutions:**

1. **Check API Service:**
   ```bash
   curl http://localhost:30080/health
   kubectl get svc kubechat-dev-api -n kubechat
   ```

2. **Check API Logs:**
   ```bash
   make dev-logs-api
   ```

3. **Debug API Container:**
   ```bash
   make dev-shell-api
   # Inside container:
   ps aux
   netstat -tlnp
   ```

### Performance Issues

#### Issue: Slow Application Response

**Symptoms:**
- Long load times
- Timeouts
- High resource usage

**Solutions:**

1. **Check Resource Usage:**
   ```bash
   kubectl top pods -n kubechat
   kubectl describe node
   ```

2. **Optimize Resources:**
   ```bash
   # Edit values-dev.yaml to increase limits
   # Restart services
   make dev-restart
   ```

3. **Clean Up:**
   ```bash
   docker system prune -f
   kubectl delete pods --field-selector=status.phase=Succeeded -n kubechat
   ```

### Network Issues

#### Issue: Port Conflicts

**Symptoms:**
- Cannot access services on expected ports
- "Port already in use" errors

**Solutions:**

1. **Check Port Usage:**
   ```bash
   lsof -i :30001  # Check web port
   lsof -i :30080  # Check API port
   lsof -i :5432   # Check PostgreSQL port
   ```

2. **Kill Conflicting Processes:**
   ```bash
   sudo lsof -ti :30001 | xargs sudo kill -9
   ```

3. **Use Port Forwarding:**
   ```bash
   make dev-port-forward
   # Access via forwarded ports instead
   ```

#### Issue: Cannot Access Services

**Symptoms:**
- Services unreachable from browser
- Connection refused errors

**Solutions:**

1. **Check Service Configuration:**
   ```bash
   kubectl get svc -n kubechat
   kubectl describe svc kubechat-dev-web -n kubechat
   ```

2. **Verify NodePort Services:**
   ```bash
   # Ensure services are configured as NodePort
   kubectl get svc -n kubechat -o wide
   ```

3. **Check Ingress (if enabled):**
   ```bash
   kubectl get ingress -n kubechat
   kubectl describe ingress kubechat-dev -n kubechat
   ```

## 🔍 Advanced Debugging

### Debug Container Issues

```bash
# Get detailed pod information
kubectl describe pod <pod-name> -n kubechat

# Check container logs
kubectl logs <pod-name> -c <container-name> -n kubechat

# Execute commands in container
kubectl exec -it <pod-name> -n kubechat -- /bin/sh

# Check container resources
kubectl top pod <pod-name> -n kubechat --containers
```

### Debug Networking

```bash
# Test pod-to-pod connectivity
kubectl exec -it <pod1> -n kubechat -- ping <pod2-ip>

# Check DNS resolution
kubectl exec -it <pod> -n kubechat -- nslookup kubernetes.default

# Test service connectivity
kubectl exec -it <pod> -n kubechat -- curl http://kubechat-dev-api:8080/health
```

### Debug Storage

```bash
# Check persistent volumes
kubectl get pv,pvc -n kubechat

# Check storage class
kubectl get storageclass

# Debug volume mounts
kubectl describe pod <pod> -n kubechat | grep -A 10 "Mounts:"
```

## 📊 Monitoring and Logs

### Application Logs

```bash
# All application logs
make dev-logs

# Specific service logs
make dev-logs-api
make dev-logs-web

# Follow logs in real-time
kubectl logs -f deployment/kubechat-dev-api -n kubechat
```

### System Events

```bash
# Recent events
kubectl get events -n kubechat --sort-by='.firstTimestamp'

# Watch events in real-time
kubectl get events -n kubechat --watch
```

### Resource Monitoring

```bash
# Resource usage (if metrics server available)
kubectl top nodes
kubectl top pods -n kubechat

# Resource requests and limits
kubectl describe nodes | grep -A 5 "Allocated resources"
```

## 🆘 Emergency Recovery

### Complete Environment Reset

If everything fails, completely reset the environment:

```bash
# 1. Stop all KubeChat services
make dev-clean

# 2. Reset Rancher Desktop
# Go to Rancher Desktop > Settings > Reset > Reset Kubernetes

# 3. Clean Docker
docker system prune -a -f

# 4. Restart Rancher Desktop

# 5. Reinitialize environment
make init
```

### Backup and Restore Data

```bash
# Backup database
./infrastructure/scripts/database/migrate.sh backup

# Restore from backup
kubectl exec -n kubechat <postgres-pod> -- psql -U kubechat -d kubechat < backup.sql
```

## 📞 Getting Help

### Before Asking for Help

1. Run diagnostics:
   ```bash
   ./infrastructure/scripts/validate-env.sh
   ./infrastructure/scripts/health-check.sh
   make dev-debug
   ```

2. Collect logs:
   ```bash
   make dev-logs > kubechat-logs.txt
   kubectl get all -n kubechat > kubechat-status.txt
   ```

3. Check existing issues on GitHub

### Creating a Bug Report

Include the following information:

- Operating system and version
- Tool versions (`make dev-info`)
- Complete error messages
- Steps to reproduce
- Environment validation output
- Relevant logs

### Support Channels

- **GitHub Issues**: For bugs and feature requests
- **GitHub Discussions**: For questions and help
- **Documentation**: Check [docs/](../docs/) for guides

---

**Remember**: KubeChat's container-first approach means issues are usually environment-related. The validation and health check scripts can resolve most problems!