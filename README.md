# Site Monitor 🛠️

A **website uptime monitoring** web app built with **Go** and **Turso (libSQL)**, following a **Test-Driven Development (TDD)** approach. The app allows users to log in, add sites for monitoring, and receive notifications via **Slack** when a site goes down.

🚀 **Just a fun side project!** Built mainly for **learning and experimenting**

## 🚀 Features

- ✅ User authentication
- 🌐 Add, edit, delete websites to monitor
- 🔄 Enable/disable monitoring for each site
- 🔔 Notifications via **Slack**

---

## 📦 Tech Stack

- **Go** (Standard Template Library for templates)
- **Turso (libSQL)** for database

---

## 🛠️ Installation & Setup

### 1️⃣ Clone the Repository

```sh
git clone https://github.com/shuvo-paul/uptimebot.git
cd uptimebot
```

### 2️⃣ Install Dependencies

Ensure you have Go installed, then run:

```sh
go mod tidy
```

### 3️⃣ Configure Environment Variables

Create a `.env` file and set the required values:

```env
Create a `.env` file and set the required values:

```env
BASE_URL=localhost:8080
TURSO_DATABASE_URL=your_turso_database_url
TURSO_AUTH_TOKEN=your_turso_auth_token

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

### 4️⃣ Run the App

```sh
pnpm install
make build
make run
```

### 5️⃣ Run Tests 🧪

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
