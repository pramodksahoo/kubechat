# KubeChat Database Disaster Recovery Guide

## Overview

This document provides comprehensive procedures for KubeChat database disaster recovery, backup management, and business continuity planning.

**Author:** James (Full Stack Developer Agent)  
**Date:** 2025-01-11  
**Version:** 1.0

## Table of Contents

1. [Emergency Contacts](#emergency-contacts)
2. [Recovery Time Objectives](#recovery-time-objectives)
3. [Backup Strategy](#backup-strategy)
4. [Recovery Procedures](#recovery-procedures)
5. [Disaster Scenarios](#disaster-scenarios)
6. [Preventive Measures](#preventive-measures)
7. [Testing Procedures](#testing-procedures)
8. [Monitoring and Alerting](#monitoring-and-alerting)

## Emergency Contacts

| Role | Contact | Phone | Email |
|------|---------|-------|-------|
| Database Administrator | TBD | TBD | dba@kubechat.dev |
| System Administrator | TBD | TBD | sysadmin@kubechat.dev |
| Security Team | TBD | TBD | security@kubechat.dev |
| Management | TBD | TBD | management@kubechat.dev |

## Recovery Time Objectives

| Scenario | RTO (Recovery Time Objective) | RPO (Recovery Point Objective) |
|----------|-------------------------------|--------------------------------|
| Database Corruption | 2 hours | 15 minutes |
| Hardware Failure | 4 hours | 15 minutes |
| Complete Site Loss | 8 hours | 1 hour |
| Data Center Outage | 12 hours | 1 hour |
| Cyber Attack | 6 hours | 30 minutes |

## Backup Strategy

### Automated Backup Schedule

```bash
# Daily full backups at 2 AM
0 2 * * * /opt/kubechat/infrastructure/scripts/database/backup.sh backup

# Hourly audit log verification
0 * * * * /opt/kubechat/infrastructure/scripts/database/health-check.sh

# Weekly cleanup of old backups
0 3 * * 0 /opt/kubechat/infrastructure/scripts/database/backup.sh cleanup
```

### Backup Types

1. **Full Database Backup**
   - Frequency: Daily
   - Retention: 30 days local, 90 days S3
   - Includes: All tables, indexes, functions, triggers

2. **Audit Log Backup**
   - Frequency: Real-time replication
   - Retention: 7 years (compliance requirement)
   - Special handling: Immutable, integrity-verified

3. **Configuration Backup**
   - Frequency: On change
   - Includes: User configs, cluster configs (encrypted)

### Backup Storage Locations

- **Primary:** Local storage (`/var/backups/kubechat`)
- **Secondary:** AWS S3 (`s3://kubechat-backups/`)
- **Tertiary:** Encrypted offsite storage (compliance)

## Recovery Procedures

### 1. Database Corruption Recovery

**Scenario:** Database corruption detected

**Steps:**

1. **Immediate Assessment**
   ```bash
   # Check database connectivity
   ./health-check.sh
   
   # Verify audit log integrity
   ./backup.sh verify latest
   ```

2. **Isolate Database**
   ```bash
   # Stop application connections
   kubectl scale deployment kubechat-api --replicas=0
   
   # Enable maintenance mode
   kubectl apply -f maintenance-mode.yaml
   ```

3. **Restore from Latest Backup**
   ```bash
   # List available backups
   ./backup.sh list
   
   # Restore latest verified backup
   ./backup.sh restore kubechat-backup-YYYYMMDD-HHMMSS.sql.gz
   ```

4. **Verify Recovery**
   ```bash
   # Run comprehensive health check
   ./health-check.sh
   
   # Verify audit log integrity
   psql -c "SELECT * FROM verify_audit_log_integrity(NULL) WHERE NOT is_valid;"
   ```

5. **Resume Operations**
   ```bash
   # Scale application back up
   kubectl scale deployment kubechat-api --replicas=3
   
   # Disable maintenance mode
   kubectl delete -f maintenance-mode.yaml
   ```

### 2. Complete Database Loss

**Scenario:** Complete database server failure

**Steps:**

1. **Provision New Database Server**
   ```bash
   # Deploy new PostgreSQL instance
   helm install postgresql bitnami/postgresql \
     --set global.postgresql.auth.postgresPassword=SECURE_PASSWORD
   ```

2. **Restore Database**
   ```bash
   # Initialize database
   psql -f init.sql
   
   # Restore latest backup
   ./backup.sh restore kubechat-backup-YYYYMMDD-HHMMSS.sql.gz
   ```

3. **Update Application Configuration**
   ```bash
   # Update database connection strings
   kubectl patch configmap kubechat-config \
     --patch '{"data":{"DATABASE_HOST":"new-db-host"}}'
   ```

### 3. Audit Log Tampering

**Scenario:** Suspected audit log manipulation

**Steps:**

1. **Immediate Investigation**
   ```bash
   # Run integrity check
   ./backup.sh verify latest
   
   # Check for integrity violations
   psql -c "SELECT * FROM verify_audit_log_integrity(NULL) WHERE NOT is_valid;"
   ```

2. **Forensic Analysis**
   ```bash
   # Export audit logs for analysis
   psql -c "COPY audit_logs TO '/tmp/audit_export.csv' WITH CSV HEADER;"
   
   # Check database logs
   tail -f /var/log/postgresql/postgresql.log
   ```

3. **Restore Clean State**
   ```bash
   # Restore from last known good backup
   ./backup.sh restore kubechat-backup-VERIFIED.sql.gz
   ```

## Disaster Scenarios

### Scenario A: Hardware Failure

**Detection:** Database server unresponsive
**Impact:** Complete service outage
**Recovery Steps:**
1. Provision replacement hardware
2. Restore from latest backup
3. Update DNS/load balancer
4. Resume service

**Estimated Recovery Time:** 4 hours

### Scenario B: Data Corruption

**Detection:** Integrity check failures
**Impact:** Partial service degradation
**Recovery Steps:**
1. Identify corruption scope
2. Restore affected tables
3. Verify data integrity
4. Resume normal operations

**Estimated Recovery Time:** 2 hours

### Scenario C: Security Breach

**Detection:** Unauthorized access alerts
**Impact:** Potential data compromise
**Recovery Steps:**
1. Immediately isolate database
2. Conduct security assessment
3. Restore from pre-breach backup
4. Implement additional security measures

**Estimated Recovery Time:** 6 hours

### Scenario D: Natural Disaster

**Detection:** Complete data center loss
**Impact:** Full system outage
**Recovery Steps:**
1. Activate disaster recovery site
2. Restore from offsite backups
3. Redirect traffic to DR site
4. Begin data center rebuild

**Estimated Recovery Time:** 12 hours

## Preventive Measures

### 1. Regular Backup Validation

```bash
# Weekly backup verification
#!/bin/bash
LATEST_BACKUP=$(ls -t /var/backups/kubechat/kubechat-backup-*.sql.gz | head -1)
./backup.sh verify "$LATEST_BACKUP"
```

### 2. Audit Log Monitoring

```bash
# Daily integrity checks
#!/bin/bash
VIOLATIONS=$(psql -t -c "SELECT COUNT(*) FROM verify_audit_log_integrity(NULL) WHERE NOT is_valid;")
if [ "$VIOLATIONS" -gt 0 ]; then
    echo "ALERT: $VIOLATIONS audit log integrity violations detected"
    # Send alert to monitoring system
fi
```

### 3. Automated Health Checks

```bash
# Continuous monitoring
*/5 * * * * /opt/kubechat/infrastructure/scripts/database/health-check.sh --quiet
```

### 4. Security Hardening

- Database user permissions limited to minimum required
- Network access restricted via firewall rules
- SSL/TLS encryption for all connections
- Regular security updates and patches

## Testing Procedures

### Monthly Backup Restore Test

1. **Setup Test Environment**
   ```bash
   # Create isolated test database
   createdb kubechat_test
   ```

2. **Restore Latest Backup**
   ```bash
   # Restore to test database
   POSTGRES_DB=kubechat_test ./backup.sh restore latest
   ```

3. **Validate Restoration**
   ```bash
   # Run health checks on test database
   POSTGRES_DB=kubechat_test ./health-check.sh
   ```

4. **Document Results**
   - Restoration time
   - Data integrity status
   - Any issues encountered

### Quarterly Disaster Recovery Drill

1. **Simulate Disaster**
   - Randomly selected scenario
   - Controlled environment
   - Time-boxed exercise

2. **Execute Recovery**
   - Follow documented procedures
   - Measure actual vs. target RTO/RPO
   - Document deviations

3. **Post-Drill Review**
   - Update procedures based on learnings
   - Address identified gaps
   - Retrain personnel if needed

## Monitoring and Alerting

### Key Metrics to Monitor

1. **Database Health**
   - Connection count
   - Query response times
   - Disk space usage
   - CPU and memory utilization

2. **Backup Status**
   - Backup completion success/failure
   - Backup file sizes and checksums
   - Backup verification status

3. **Audit Trail Integrity**
   - Checksum validation results
   - Suspicious activity patterns
   - Failed authentication attempts

### Alert Thresholds

| Metric | Warning | Critical |
|--------|---------|----------|
| Disk Space | 80% | 90% |
| Connection Count | 70% of max | 90% of max |
| Query Response Time | >500ms | >2000ms |
| Failed Backups | 1 failure | 2 consecutive failures |
| Integrity Violations | 1 violation | >1 violation |

### Alert Notifications

- **Email:** For non-critical alerts during business hours
- **SMS/Phone:** For critical alerts requiring immediate attention
- **Slack/Teams:** For team coordination and status updates
- **Incident Management:** For formal incident tracking

## Recovery Success Criteria

### Technical Validation

- [ ] Database connectivity restored
- [ ] All required tables present and accessible
- [ ] Audit log integrity verified
- [ ] Application functionality confirmed
- [ ] Performance within acceptable thresholds

### Business Validation

- [ ] User authentication working
- [ ] Critical business processes functional
- [ ] Data consistency verified
- [ ] Security controls operational
- [ ] Compliance requirements met

## Contact Information for Vendors

| Service | Contact | Support Level |
|---------|---------|---------------|
| AWS Support | +1-XXX-XXX-XXXX | Business |
| PostgreSQL Community | community.postgresql.org | Community |
| Kubernetes Support | support.kubernetes.io | Community |

## Document Maintenance

- **Review Frequency:** Quarterly
- **Update Triggers:** Infrastructure changes, new threats, lessons learned
- **Approval Required:** Database Administrator, Security Team
- **Distribution:** All technical staff, management

---

**Remember:** In a disaster scenario, stay calm, follow procedures, and prioritize data integrity over speed.

For emergency assistance, contact the on-call engineer immediately.

**Last Updated:** 2025-01-11  
**Next Review Date:** 2025-04-11