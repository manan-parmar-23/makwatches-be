# Dynamic Environment Variable Management

## Overview
This system automatically syncs ALL environment variables from `.env` to `docker-compose.prod.yml`, eliminating the need to manually update the compose file when adding new variables.

## How It Works

1. **Edit `.env`** - Add/modify any environment variable
2. **Run sync script** - Automatically updates docker-compose.prod.yml
3. **Deploy** - Push changes and redeploy

## Quick Start

### Option 1: One Command Deployment (RECOMMENDED) 🚀
```bash
make deploy-all
```
This single command:
- ✅ Syncs all env variables to docker-compose.prod.yml
- ✅ Updates production environment
- ✅ Commits changes
- ✅ Pushes to GitHub (triggers CI/CD)
- ✅ Waits for CI/CD
- ✅ Redeploys application
- ✅ Verifies deployment

### Option 2: Step-by-Step
```bash
# 1. Sync env variables to compose file
make sync-env

# 2. Commit changes
git add .env docker-compose.prod.yml
git commit -m "Update environment variables"
git push

# 3. Deploy
make deploy-prod
```

### Option 3: Manual Sync Only
```bash
# Just sync without deploying
./sync-env-to-compose.sh
```

## Adding New Environment Variables

### Before (Old Way) ❌
1. Edit `.env`
2. Manually add variable to `docker-compose.prod.yml` under `environment:` section
3. Remember exact formatting
4. Commit both files
5. Deploy

### Now (New Way) ✅
1. Edit `.env` and add your variable:
   ```env
   NEW_FEATURE_API_KEY=your_api_key_here
   ```
2. Run ONE command:
   ```bash
   make deploy-all
   ```
3. Done! 🎉

## Example Workflow

### Scenario: Adding a new API integration

```bash
# 1. Add the new API credentials to .env
echo "STRIPE_API_KEY=sk_live_xxxxxxxxx" >> .env
echo "STRIPE_WEBHOOK_SECRET=whsec_xxxxxxxx" >> .env

# 2. Deploy everything
make deploy-all
# Enter commit message when prompted: "Add Stripe payment integration"

# 3. That's it! The new variables are:
#    - In docker-compose.prod.yml ✓
#    - In production .env ✓
#    - Available in the container ✓
```

## Available Commands

| Command | Description |
|---------|-------------|
| `make deploy-all` | **ONE-COMMAND DEPLOY**: Sync, commit, push, deploy everything |
| `make sync-env` | Sync .env variables to docker-compose.prod.yml |
| `make update-prod-env` | Update production .env file |
| `make deploy-prod` | Redeploy production containers |
| `./deploy-with-env.sh` | Full deployment script (same as make deploy-all) |
| `./sync-env-to-compose.sh` | Just sync env to compose |

## Technical Details

### What Gets Synced?
- **ALL** variables from `.env` (except comments and empty lines)
- Automatically sorted alphabetically
- Proper docker-compose format

### File Structure
```
makwatches-be/
├── .env                          # Source of truth for env variables
├── docker-compose.prod.yml       # Auto-generated from .env
├── sync-env-to-compose.sh        # Sync script
├── deploy-with-env.sh            # One-command deployment
└── update-production-env.sh      # Sync to /opt/makwatches-be
```

### How Sync Works
1. Reads all variables from `.env` (ignoring comments)
2. Generates docker-compose environment section dynamically
3. Creates backup of current docker-compose.prod.yml
4. Writes new docker-compose.prod.yml with all variables

### Safety Features
- ✅ Automatic backups before overwriting
- ✅ Validates .env file exists
- ✅ Shows count and list of synced variables
- ✅ Git commit protection (won't overwrite uncommitted changes)

## Current Environment Variables

Your `.env` currently has these variables (auto-synced):
- DOCKER_USERNAME
- ENVIRONMENT
- FIREBASE_BUCKET_NAME
- FIREBASE_CREDENTIALS_PATH
- FIREBASE_PROJECT_ID
- GOOGLE_CLIENT_ID
- GOOGLE_CLIENT_SECRET
- GOOGLE_REDIRECT_URL
- JWT_EXPIRY
- JWT_SECRET
- MONGO_URI
- PORT
- RAZORPAY_KEY_ID
- RAZORPAY_KEY_ID_TEST
- RAZORPAY_KEY_SECRET
- RAZORPAY_KEY_SECRET_TEST
- RAZORPAY_MODE
- REDIS_DATABASE_NAME
- REDIS_PASSWORD
- REDIS_URI

## Troubleshooting

### Variables not appearing in container
```bash
# 1. Ensure sync ran
make sync-env

# 2. Check docker-compose.prod.yml
cat docker-compose.prod.yml | grep -A 30 "environment:"

# 3. Recreate containers (not just restart)
make deploy-prod
```

### Backup files piling up
```bash
# Clean old backups (keeps last 5)
find . -name "docker-compose.prod.yml.backup.*" -type f | sort -r | tail -n +6 | xargs rm -f
```

### Need to rollback
```bash
# List backups
ls -lt docker-compose.prod.yml.backup.*

# Restore from backup
cp docker-compose.prod.yml.backup.YYYYMMDD_HHMMSS docker-compose.prod.yml
```

## Best Practices

1. **Always use `make deploy-all`** for env changes
2. **Test in development first** before deploying to production
3. **Keep `.env` in `.gitignore`** (it's already there)
4. **Use descriptive commit messages** when prompted
5. **Verify deployment** after running deploy-all

## Migration from Old System

If you have existing variables manually added to docker-compose.prod.yml:

```bash
# 1. Ensure all variables are in .env
cat docker-compose.prod.yml | grep "- " | sed 's/.*- \([^=]*\)=.*/\1/' > current_vars.txt

# 2. Check if they're in .env
while read var; do grep -q "^$var=" .env || echo "Missing: $var"; done < current_vars.txt

# 3. Add any missing variables to .env

# 4. Run sync
make sync-env

# 5. Verify
diff <(cat docker-compose.prod.yml | grep "- " | sort) <(cat docker-compose.prod.yml.backup.* | head -1 | grep "- " | sort)
```

## Security Notes

- ⚠️ Never commit `.env` to Git (already in `.gitignore`)
- ⚠️ Keep production `.env` secure
- ⚠️ Use different credentials for dev/staging/prod
- ⚠️ Rotate sensitive keys regularly

## Summary

**Before**: Manual env management, error-prone, multiple files to edit
**Now**: Edit `.env` → Run `make deploy-all` → Done! ✨

Questions? Check the logs:
```bash
docker logs makwatches-be-api --tail=50
```
