English | [简体中文](./README.zh-CN.md)

# ZbxTable

[![Build and Test](https://github.com/canghai908/zbxtable/actions/workflows/build.yml/badge.svg)](https://github.com/canghai908/zbxtable/actions/workflows/build.yml)
[![Release](https://github.com/canghai908/zbxtable/actions/workflows/release.yml/badge.svg)](https://github.com/canghai908/zbxtable/actions/workflows/release.yml)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)
[![Go Version](https://img.shields.io/badge/Go-%3E%3D%201.23-blue)](https://golang.org/)

ZbxTable is an intelligent Zabbix operations platform developed using Go language, providing powerful monitoring data visualization, alert management, and AI-powered intelligent analysis.

## ✨ What's New in 3.0

**ZbxTable 3.0** is a milestone release with major feature enhancements:

- 🤖 **AI-Powered Analysis** - Integrated AI models (Ollama/Deepseek) for intelligent alert analysis and solution recommendations
- 🌐 **Multi-Tenant Architecture** - Support for managing multiple Zabbix instances with instance-level data isolation and access control
- 📢 **Enhanced Alert Management** - New alert distribution rule engine with alert suppression, status tracking, and event logging
- 💾 **Database Optimization** - Full support for MySQL and PostgreSQL with unified GORM data access layer
- 🔐 **Enhanced Security** - Encrypted storage for sensitive information with auto-generated encryption keys
- ⚡ **Cache System Refactoring** - Using go-cache instead of Redis for simplified deployment
- 📝 **Upgraded Logging System** - Automatic log rotation by date with compression of historical logs

## Features

### Core Features
- 🎨 **Topology Management** - Custom topology drawing with real-time status display
- 📊 **Device Management** - Device classification and export (Linux, Windows, Network, Storage, etc.)
- 📈 **Report System** - Export Zabbix alerts and metrics data to Excel for specific time periods
- 🔍 **Alert Analysis** - Analyze alert data for specific time periods, generate Top 10 alerts, etc.
- 📅 **Scheduled Tasks** - Scheduled report generation and automatic delivery

### Alert Management
- 📧 **Multi-Channel Notifications** - Support for Email, WeChat Work, Webhook, and more
- 🎯 **Alert Distribution** - Flexible alert rule engine with multi-dimensional matching conditions
- 🔕 **Alert Suppression** - Multi-dimensional suppression rules (time period, host, severity, etc.)
- 📊 **Alert Tracking** - Complete alert status tracking and event logging
- 🤖 **AI Analysis** - Intelligent alert cause analysis with solution recommendations

### System Management
- 🌐 **Multi-Instance Management** - Manage multiple Zabbix instances from a unified platform
- 🔐 **Access Control** - Role-based access control (RBAC) with fine-grained permissions
- ⚙️ **Configuration Management** - Database-driven configuration system with runtime updates
- 📱 **Menu Management** - Dynamic menu system with role-based menu permissions
- 🔒 **Security Encryption** - Automatic encryption for sensitive information

## Architecture

![1](/zbxtable.png)

## Component

**ZbxTable**: Backend written using [Gin framework](https://github.com/gin-gonic/gin) and Go language.

**ZbxTable-Web**: Front end written using [Vue](https://github.com/vuejs/vue).

## 🛠️ Technology Stack

### Backend
- **Language**: Go 1.23+
- **Web Framework**: Gin
- **ORM**: GORM
- **Database**: MySQL 5.7+ / PostgreSQL 12+
- **AI Integration**: Ollama / Deepseek API

### Frontend
- **Framework**: Vue.js 2.x
- **UI Components**: Ant Design Vue
- **Charts**: ECharts
- **Topology**: G6

### Alert Reception
- **Webhook Method**: ZbxTable provides Webhook API endpoint to receive Zabbix alerts directly
- **Endpoint**: `http://your-zbxtable-server:8088/v1/receive`
- **Authentication**: Token-based authentication
- **Format**: JSON

Configure Webhook media type in Zabbix to send alerts to ZbxTable's Webhook endpoint for alert reception and distribution. For detailed configuration, please refer to the [official documentation](https://zbxtable.com).

## Demo

[https://demo.zbxtable.com](https://demo.zbxtable.com)

**Default credentials:** `admin / Zbxtable`

## Compatibility

| Zabbix Version | Compatibility |
|:---------------|:-------------:|
| 7.4.x          |       ✅       |
| 7.2.x          |       ✅       |
| 7.0.x LTS      |       ✅       |
| 6.4.x          |       ✅       |
| 6.2.x          |       ✅       |
| 6.0.x LTS      |       ✅       |
| 5.4.x          |       ✅       |
| 5.2.x          |       ✅       |
| 5.0.x LTS      |       ✅       |
| 4.4.x          |       ✅       |
| 4.2.x          |       ✅       |
| 4.0.x LTS      |       ✅       |
| 3.4.x          |   untested    |
| 3.2.x          |   untested    |
| 3.0.x LTS      |   untested    |

## ⚠️ Important Notice

**ZbxTable 3.0 is not compatible with 2.x database structure. Fresh installation is recommended.** If you are using version 2.x, please backup important data and perform a fresh installation of version 3.0.

## Quick Start

### Method 1: One-Click Installation Script (Recommended)

Suitable for CentOS, Ubuntu, Debian and other Linux systems with automatic configuration.

```bash
# Online installation
curl -fsSL https://dl.cactifans.com/zbxtable/install.sh | sudo bash

# Or download and install
wget https://dl.cactifans.com/zbxtable/install.sh
sudo bash install.sh
```

**Installation script features:**
- ✅ Automatic OS detection
- ✅ Download latest version automatically
- ✅ Create system user and directories
- ✅ Configure systemd service
- ✅ Configure firewall rules
- ✅ Start service automatically

**After installation:**
- First installation: `http://your-server-ip:8088`
- Default credentials: `admin / Zbxtable` (Please change immediately)

### Method 2: Binary Deployment

#### 1. Download Binary

Download from GitHub Releases or mirror site:

```bash
# Download from GitHub (e.g., v3.0.0)
wget https://github.com/canghai908/zbxtable/releases/download/v3.0.0/zbxtable-v3.0.0-linux-amd64.tar.gz

# Or download from mirror site
wget https://dl.cactifans.com/zbxtable/zbxtable-latest-linux-amd64.tar.gz

# Extract
tar -xzf zbxtable-v3.0.0-linux-amd64.tar.gz
cd zbxtable-v3.0.0-linux-amd64
```

#### 2. Create Database

**MySQL:**

```sql
CREATE DATABASE zbxtable CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
CREATE USER 'zbxtable'@'localhost' IDENTIFIED BY 'your_password';
GRANT ALL PRIVILEGES ON zbxtable.* TO 'zbxtable'@'localhost';
FLUSH PRIVILEGES;
```

**PostgreSQL:**

```sql
CREATE DATABASE zbxtable;
CREATE USER zbxtable WITH PASSWORD 'your_password';
GRANT ALL PRIVILEGES ON DATABASE zbxtable TO zbxtable;
```

#### 3. Start Service

**Create System User:**

```bash
# Create zbxtable system user
sudo useradd -r -s /sbin/nologin zbxtable

# Create installation directory
sudo mkdir -p /usr/local/zbxtable

# Copy binary file
sudo cp zbxtable /usr/local/zbxtable/

# Set permissions
sudo chown -R zbxtable:zbxtable /usr/local/zbxtable
sudo chmod +x /usr/local/zbxtable/zbxtable
```

**Configure systemd Service:**

```bash
# Create systemd service file
sudo tee /etc/systemd/system/zbxtable.service > /dev/null << EOF
[Unit]
Description=ZbxTable - Zabbix Monitoring Tool
After=network.target

[Service]
Type=simple
User=zbxtable
Group=zbxtable
WorkingDirectory=/usr/local/zbxtable
ExecStart=/usr/local/zbxtable/zbxtable web
Restart=on-failure
RestartSec=5s

[Install]
WantedBy=multi-user.target
EOF

# Reload systemd
sudo systemctl daemon-reload

# Start service
sudo systemctl start zbxtable

# Enable auto-start on boot
sudo systemctl enable zbxtable

# Check service status
sudo systemctl status zbxtable
```

#### 4. Access Installation Wizard

Visit `http://your-server-ip:8088` and follow the installation wizard to configure database connection and complete the setup.

## 📚 Documentation

- [Release Notes 3.0](./RELEASE_3.0.md) - Detailed version update information
- [System Documentation](./docs/SYSTEM.md) - Complete system documentation
- [API Documentation](./docs/api/README.md) - API reference
- [Official Website](https://zbxtable.com) - More documentation and tutorials

## 📊 System Requirements

### Minimum Configuration

- **CPU**: 2 cores
- **Memory**: 4 GB
- **Disk**: 20 GB
- **OS**: CentOS 7/8, Ubuntu 18.04+, Debian 9+, RHEL 7/8

### Recommended Configuration

- **CPU**: 4 cores
- **Memory**: 8 GB
- **Disk**: 50 GB SSD
- **Database**: MySQL 5.7+ or PostgreSQL 12+

### Network Requirements

- Access to Zabbix Server API
- Access to SMTP server (if using email notifications)
- Access to WeChat Work API (if using WeChat notifications)
- Internet access (if using Deepseek API)

## 🔄 Upgrading from 2.1 to 3.0

**⚠️ Important: ZbxTable 3.0 is completely incompatible with 2.x database structure and cannot be directly upgraded.**

Recommended steps:

1. **Backup important data from 2.x version** (topology maps, custom configurations, etc.)
2. **Perform fresh installation of 3.0 version** (using one-click script or binary deployment)
3. **Reconfigure the system** (Zabbix instances, alert rules, etc.)

If you must retain historical data, please contact technical support for data migration solutions.

## 🔒 Security Recommendations

1. **Change Default Password** - Change admin password immediately after first login
2. **Use HTTPS** - Configure Nginx reverse proxy with SSL certificate
3. **Firewall Configuration** - Only open necessary ports and restrict source IPs
4. **Regular Backups** - Backup database and configuration files daily
5. **Permission Management** - Use principle of least privilege and review user permissions regularly

## 📦 Code Repositories

**ZbxTable**: [https://github.com/canghai908/zbxtable](https://github.com/canghai908/zbxtable)

**ZbxTable-Web**: [https://github.com/canghai908/zbxtable-web](https://github.com/canghai908/zbxtable-web)

## 👥 Team

**Back-end development**: [canghai908](https://github.com/canghai908)

**Front-end development**: [ahyiru](https://github.com/ahyiru)

## 🤝 Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## 📞 Support

- **GitHub Issues**: [https://github.com/canghai908/zbxtable/issues](https://github.com/canghai908/zbxtable/issues)
- **Documentation**: [https://zbxtable.com](https://zbxtable.com)
- **Demo**: [https://demo.zbxtable.com](https://demo.zbxtable.com)

## 📄 License

ZbxTable is available under the Apache-2.0 license. See the [LICENSE](LICENSE) file for more info.

---

**ZbxTable 3.0 - Making Zabbix Monitoring Smarter and More Powerful!** 🚀
