# UAT Deployment Guide — Visionary MVP

**Stack:** NestJS Backend → GCP VM (existing) | Next.js Frontend → Vercel
**Trigger:** Push / Merge to `main` branch → GitHub Actions → auto deploy

---

## Architecture Overview

```
GitHub (main branch)
        │
        ├── GitHub Actions (CI/CD)
        │         │
        │    ┌────┴────────────────────┐
        │    │                         │
        │  SSH into GCP VM         Vercel CLI
        │  git pull + build        auto deploy
        │  pm2 restart             Next.js frontend
        │    │
        │  GCP VM (NestJS API)
        │  Nginx reverse proxy
        │  PM2 process manager
        │
        └── Supabase UAT Project (separate DB)
```

---

## Part 1 — First Time VM Setup (Do this once)

You already have the VM running. Connect via SSH terminal (like in FileZilla or use GCP Console SSH).

---

### Step 1.1 — Connect to the VM

In GCP Console → **Compute Engine → VM instances → SSH button**
Or use your existing SSH client (FileZilla terminal).

```bash
# You should see this prompt after connecting:
ashish@uat-instance-20260417-072529:~$
```

---

### Step 1.2 — Install Node.js 20

```bash
# Install nvm (Node version manager)
curl -o- https://raw.githubusercontent.com/nvm-sh/nvm/v0.39.7/install.sh | bash

# Reload shell
source ~/.bashrc

# Install Node.js 20
nvm install 20
nvm use 20
nvm alias default 20

# Verify
node -v   # should show v20.x.x
npm -v
```

---

### Step 1.3 — Install PM2 (Process Manager)

PM2 keeps your NestJS app running and auto-restarts it if it crashes.

```bash
npm install -g pm2

# Verify
pm2 -v
```

---

### Step 1.4 — Install Nginx (Reverse Proxy)

```bash
sudo apt update
sudo apt install nginx -y

# Start and enable on boot
sudo systemctl start nginx
sudo systemctl enable nginx

# Check it's running
sudo systemctl status nginx
```

---

### Step 1.5 — Install Git

```bash
sudo apt install git -y
git --version
```

---

### Step 1.6 — Clone the Repository

```bash
# Go to home directory
cd ~

# Clone your repo (use HTTPS for now)
git clone https://github.com/visionaryorg/visionary-dev.git

# Go into the backend folder
cd visionary-dev/backend
```

---

### Step 1.7 — Create the .env File on the VM

```bash
# Create the env file (never commit this)
nano ~/visionary-dev/backend/.env
```

Paste your environment variables:

```env
NODE_ENV=production
PORT=3001

DATABASE_URL=your-supabase-uat-pooler-url
DIRECT_URL=your-supabase-uat-direct-url

JWT_SECRET=your-jwt-secret
JWT_EXPIRES_IN=7d

GOOGLE_CLIENT_ID=your-google-client-id
GOOGLE_CLIENT_SECRET=your-google-client-secret
GOOGLE_CALLBACK_URL=http://YOUR-VM-IP/api/auth/google/callback

SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
SMTP_SECURE=false
SMTP_USER=your-email@gmail.com
SMTP_PASS=your-app-password
SMTP_FROM=Visionary <your-email@gmail.com>

FRONTEND_URL=https://your-app.vercel.app
```

Save: `Ctrl+X` → `Y` → `Enter`

---

### Step 1.8 — Install Dependencies & Build

```bash
cd ~/visionary-dev/backend

# Install dependencies
npm install

# Generate Prisma client
npx prisma generate

# Run migrations
npx prisma migrate deploy

# Build the NestJS app
npm run build
```

---

### Step 1.9 — Start with PM2

```bash
cd ~/visionary-dev/backend

# Start the app with PM2
pm2 start dist/src/main.js --name visionary-backend

# Save PM2 process list (survives VM reboot)
pm2 save

# Set PM2 to auto-start on VM reboot
pm2 startup
# ↑ It will print this command — copy and run it:
sudo env PATH=$PATH:/usr/bin /usr/lib/node_modules/pm2/bin/pm2 startup systemd -u ashish --hp /home/ashish

# Check it's running
pm2 status
pm2 logs visionary-backend
```

---

### Step 1.10 — Configure Nginx as Reverse Proxy

```bash
# Create Nginx config for the backend
sudo nano /etc/nginx/sites-available/visionary-backend
```

Paste this config:

```nginx
server {
    listen 80;
    server_name _;

    location /api/ {
        proxy_pass http://localhost:3001/api/;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection 'upgrade';
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_cache_bypass $http_upgrade;
    }
}
```

Save: `Ctrl+X` → `Y` → `Enter`

```bash
# Enable the config
sudo ln -s /etc/nginx/sites-available/visionary-backend /etc/nginx/sites-enabled/

# Remove default nginx page
sudo rm /etc/nginx/sites-enabled/default

# Test config is valid
sudo nginx -t

# Reload Nginx
sudo systemctl reload nginx
```

---

### Step 1.10b — Configure Nginx for Subdomain (if using a domain like uat-demo1.visionary.org.in)

If your team has set up a subdomain with Certbot/SSL, the auto-generated config only serves static files. You need to add the `/api/` proxy block manually.

```bash
sudo nano /etc/nginx/sites-available/uat-demo1.visionary.org.in
```

Find the `location / {` block inside the `server` (port 443/SSL) block and **add the `/api/` block above it**:

```nginx
server {
        server_name uat-demo1.visionary.org.in;

        root /var/www/react-app;
        index index.html index.htm;

        location /api/ {
                proxy_pass http://localhost:3001/api/;
                proxy_http_version 1.1;
                proxy_set_header Upgrade $http_upgrade;
                proxy_set_header Connection 'upgrade';
                proxy_set_header Host $host;
                proxy_set_header X-Real-IP $remote_addr;
                proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
                proxy_cache_bypass $http_upgrade;
        }

        location / {
                try_files $uri $uri/ /index.html;
        }

    listen 443 ssl; # managed by Certbot
    listen [::]:443 ssl ipv6only=on; # managed by Certbot
    ssl_certificate /etc/letsencrypt/live/uat-demo1.visionary.org.in/fullchain.pem; # managed by Certbot
    ssl_certificate_key /etc/letsencrypt/live/uat-demo1.visionary.org.in/privkey.pem; # managed by Certbot
    include /etc/letsencrypt/options-ssl-nginx.conf; # managed by Certbot
    ssl_dhparam /etc/letsencrypt/ssl-dhparams.pem; # managed by Certbot
}
```

Save: `Ctrl+X` → `Y` → `Enter`

```bash
sudo nginx -t
sudo systemctl reload nginx
```

> **Also update your backend `.env`** on the VM:
> ```
> GOOGLE_CALLBACK_URL=https://uat-demo1.visionary.org.in/api/auth/google/callback
> FRONTEND_URL=https://visionary-dev-gamma.vercel.app
> ```
> Then: `pm2 restart visionary-backend`

---

### Step 1.11 — Open Port 80 in GCP Firewall

1. Go to GCP Console → **VPC Network → Firewall**
2. Click **Create Firewall Rule**
3. Fill in:
   - **Name:** `allow-http`
   - **Targets:** All instances in the network
   - **Source IP ranges:** `0.0.0.0/0`
   - **Protocols and ports:** TCP → `80`
4. Click **Create**

Now test: open `http://YOUR-VM-EXTERNAL-IP/api/grades` in your browser — you should get a JSON response.

---

## Part 2 — Vercel Setup (Frontend)

### Step 2.1 — Create Vercel Account & Link GitHub

1. Go to [vercel.com](https://vercel.com) → Sign up with GitHub
2. Click **Add New Project**
3. Select your GitHub repo `visionaryorg/visionary_mvp_ai`
4. Set **Root Directory** to `frontend`
5. Framework will auto-detect as **Next.js**
6. Click **Deploy**

---

### Step 2.2 — Configure Environment Variables in Vercel

In Vercel project → **Settings → Environment Variables**, add:

| Key | Value |
|-----|-------|
| `NEXT_PUBLIC_API_URL` | `http://YOUR-VM-EXTERNAL-IP` |

> Replace `YOUR-VM-EXTERNAL-IP` with your actual GCP VM external IP.
> Find it in GCP Console → Compute Engine → VM instances → External IP column.

---

### Step 2.3 — Set Vercel Production Branch

In Vercel project → **Settings → Git**:
- **Production Branch:** `main`

---

## Part 3 — GitHub Actions CI/CD

### Step 3.1 — Create SSH Key for GitHub Actions

On your **local machine** (not the VM), generate a deploy SSH key:

```bash
ssh-keygen -t ed25519 -C "github-actions-deploy" -f ~/.ssh/github_actions_deploy
# Press Enter for no passphrase
```

This creates two files:
- `~/.ssh/github_actions_deploy` — **private key** (goes to GitHub Secrets)
- `~/.ssh/github_actions_deploy.pub` — **public key** (goes to VM)

---

### Step 3.2 — Add Public Key to VM

Copy the public key content:

```bash
cat ~/.ssh/github_actions_deploy.pub
```

SSH into the VM and add it:

```bash
# On the VM
mkdir -p ~/.ssh
nano ~/.ssh/authorized_keys
```

Paste the public key at the end of the file. Save and exit.

```bash
# Set correct permissions
chmod 600 ~/.ssh/authorized_keys
chmod 700 ~/.ssh
```

---

### Step 3.3 — Add GitHub Secrets

Go to GitHub repo → **Settings → Secrets and variables → Actions → New repository secret**

| Secret Name | Value |
|-------------|-------|
| `VM_HOST` | Your VM external IP (e.g. `34.131.224.54`) |
| `VM_USER` | Your VM username (e.g. `ashish`) |
| `VM_SSH_KEY` | Content of `~/.ssh/github_actions_deploy` (private key — the whole thing including `-----BEGIN...-----END`) |
| `VERCEL_TOKEN` | From Vercel → Settings → Tokens → Create |
| `VERCEL_ORG_ID` | See Step 3.5 below |
| `VERCEL_PROJECT_ID` | See Step 3.5 below |

---

### Step 3.4 — Create GitHub Actions Workflow

Create the file `.github/workflows/deploy.yml` in your repo root:

```yaml
name: Deploy to UAT

on:
  push:
    branches:
      - main

jobs:
  # ── Backend: SSH Deploy to GCP VM ─────────────────────────────
  deploy-backend:
    name: Deploy Backend to GCP VM
    runs-on: ubuntu-latest

    steps:
      - name: Deploy via SSH
        uses: appleboy/ssh-action@v1.0.3
        with:
          host: ${{ secrets.VM_HOST }}
          username: ${{ secrets.VM_USER }}
          key: ${{ secrets.VM_SSH_KEY }}
          script: |
            cd ~/visionary-dev

            # Pull latest code
            git pull origin main

            # Go to backend
            cd backend

            # Install dependencies
            npm install

            # Regenerate Prisma client
            npx prisma generate

            # Run any new migrations
            npx prisma migrate deploy

            # Build
            npm run build

            # Restart with PM2
            pm2 restart visionary-backend || pm2 start dist/src/main.js --name visionary-backend

            # Save PM2 state
            pm2 save

            echo "✅ Backend deployed successfully"

  # ── Frontend: Deploy to Vercel ────────────────────────────────
  deploy-frontend:
    name: Deploy Frontend to Vercel
    runs-on: ubuntu-latest
    needs: deploy-backend

    steps:
      - name: Checkout code
        uses: actions/checkout@v4

      - name: Install Vercel CLI
        run: npm install -g vercel@latest

      - name: Deploy to Vercel
        run: |
          cd frontend
          vercel deploy --prod \
            --token=${{ secrets.VERCEL_TOKEN }} \
            --yes
        env:
          VERCEL_ORG_ID: ${{ secrets.VERCEL_ORG_ID }}
          VERCEL_PROJECT_ID: ${{ secrets.VERCEL_PROJECT_ID }}
```

---

### Step 3.5 — Get Vercel IDs

```bash
# Install Vercel CLI locally
npm i -g vercel

# Login
vercel login

# Inside the frontend folder
cd frontend
vercel link

# Open the generated file
cat .vercel/project.json
# orgId      → VERCEL_ORG_ID
# projectId  → VERCEL_PROJECT_ID
```

---

## Part 4 — Supabase UAT Project

Create a **separate** Supabase project for UAT so it doesn't touch production data.

1. Go to [supabase.com](https://supabase.com) → New Project → Name: `visionary-uat`
2. **Settings → Database → Connection string**
3. Copy **Transaction pooler** → paste as `DATABASE_URL` in VM's `.env`
4. Copy **Direct connection** → paste as `DIRECT_URL` in VM's `.env`

> After updating `.env` on the VM, run `pm2 restart visionary-backend`

---

## Part 5 — First Deploy (After Setup)

```bash
# Trigger CI/CD by pushing to main
git add -A
git commit -m "trigger: initial UAT deploy"
git push origin main
```

Go to GitHub → **Actions** tab to watch it live.

---

## Day-to-Day Workflow

```
Developer merges PR → main
        ↓
GitHub Actions triggers
        ↓
    ┌───┴───────────────────┐
    │                       │
SSH into VM              Vercel deploys
git pull + build         frontend
pm2 restart              automatically
    │
~3-5 minutes total
        ↓
✅ Backend live on GCP VM
✅ Frontend live on Vercel
✅ Migrations run automatically
```

---

## Useful Commands on the VM

```bash
# Check backend status
pm2 status

# View live logs
pm2 logs visionary-backend

# Restart backend manually
pm2 restart visionary-backend

# Check Nginx status
sudo systemctl status nginx

# View Nginx error logs
sudo tail -f /var/log/nginx/error.log

# Check what's running on port 3001
sudo lsof -i :3001
```

---

## Troubleshooting

### Backend not responding after deploy
```bash
# SSH into VM and check logs
pm2 logs visionary-backend --lines 50
```

### Prisma migration fails
- Make sure `DIRECT_URL` in VM `.env` uses the **direct connection** (port 5432, not pooler)

### CORS errors on frontend
In `backend/src/main.ts`, make sure CORS allows your Vercel domain:
```typescript
app.enableCors({
  origin: ['https://your-app.vercel.app', 'http://localhost:3000'],
  credentials: true,
});
```

### Port 80 not accessible
- Check GCP Firewall rule allows TCP port 80
- Check Nginx is running: `sudo systemctl status nginx`
- Check backend is running: `pm2 status`

### VM reboots and backend stops
```bash
# Run this once on the VM to auto-start PM2 on reboot
pm2 startup
# Copy and run the command it prints
pm2 save
```

