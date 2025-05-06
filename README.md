# Site Monitor 🛠️

A **website uptime monitoring** web app built with **Go** and **PostgreSQL**, following a **Test-Driven Development (TDD)** approach. The app allows users to log in, add sites for monitoring, and receive notifications via **Slack** when a site goes down.

🚀 **Just a fun side project!** Built mainly for **learning and experimenting**

## 🚀 Features

- ✅ User authentication
- 🌐 Add, edit, delete websites to monitor
- 🔄 Enable/disable monitoring for each site
- 🔔 Notifications via **Slack**

---

## 📦 Tech Stack

- **Go** (Standard Template Library for templates)
- **PostgreSQL** for database

---

## 🛠️ Installation & Setup

### 1️⃣ Clone the Repository

```sh
git clone https://github.com/shuvo-paul/uptimebot.git
cd uptimebot
```
### 2️⃣ Configure Environment Variables

Create a `.env` file and set the required values:

```env
BASE_URL=localhost:8080
PORT=8080 # Default port is 8080

# Database Configuration
DB_HOST=localhost
DB_PORT=5432
DB_USER=uptimebot
DB_PASSWORD=uptimebot
DB_NAME=uptimebot
DB_SSLMODE=disable

# Slack OAuth Credentials
SLACK_CLIENT_ID=
SLACK_CLIENT_SECRET=
SLACK_REDIRECT_URI=https://localhost:8080/targets/auth/slack/callback

# SMTP Email Configuration
SMTP_HOST=
SMTP_PORT=587
SMTP_USERNAME=
SMTP_PASSWORD=
SMTP_EMAIL_FROM=
```

### 3️⃣ Run the App

To run the application using Docker Compose, execute the following command:

```sh
docker compose up
```

This will build and start the application along with the PostgreSQL database in a Docker container. Ensure that your `.env` file is correctly configured with the necessary environment variables.

### 4️⃣ Run Tests 🧪

```sh
go test ./...
```

---

## 🚀 Usage

1. **Login/Register**
2. **Add website URLs for monitoring**
3. **Enable/Disable monitoring**
4. **Integrate Slack to receive notifications**
5. **Get notified when a site goes down (via Slack)**

---

## 🛠️ Roadmap

- [ ] Add **Email** notifications
- [ ] Add **SMS** notifications
