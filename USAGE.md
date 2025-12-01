# tnmanage - TrueNAS Dataset Management Tool

A command-line tool to manage datasets and NFS shares on TrueNAS.

## Setup

### Option 1: Using config command (recommended)

```bash
# Set the server URL
tnmanage config server https://10.0.5.30

# Set the API token
tnmanage config token your-api-token
```

Configuration is saved to `~/.tnmanage` and automatically loaded for all commands.

### Option 2: Using environment variables

```bash
export TRUENAS_URL="https://your-truenas-server"
export TRUENAS_API_KEY="your-api-key"
```

### Option 3: Using command-line flags

Pass `--server` and `--token` flags to individual commands.

## Build

```bash
make build
```

Or install to `/usr/local/bin`:

```bash
make install
```

## Usage

### Add a new dataset

```bash
# Create a dataset with 100GB quota
tnmanage add mypool mydataset 100 --server https://10.0.5.30 --token your-api-token

# Create a dataset with NFS share for specific hosts
tnmanage add mypool mydataset 100 --server https://10.0.5.30 --token your-api-token --nfs 10.0.5.0/24,192.168.1.100

# Using environment variables
export TRUENAS_URL="https://10.0.5.30"
export TRUENAS_API_KEY="your-api-token"
tnmanage add mypool mydataset 100 --nfs 10.0.5.0/24
```

The `--nfs` flag automatically creates an NFS share with:
- maproot user: root
- maproot group: wheel
- Authorized hosts: as specified in the flag

### List all datasets in a pool

```bash
# Using environment variables
tnmanage list mypool

# Using flags
tnmanage list mypool --server https://10.0.5.30 --token your-api-token
```

### Remove a dataset

```bash
# Remove a dataset (with confirmation prompt)
tnmanage remove mypool/mydataset

# Skip confirmation prompt with --force flag
tnmanage remove mypool/mydataset --force

# Using flags
tnmanage remove mypool/mydataset --server https://10.0.5.30 --token your-api-token
```

**Warning:** This operation permanently deletes the dataset and all its data. This cannot be undone!

### Clear/wipe a dataset

```bash
# Clear all data from a dataset (with confirmation prompt)
tnmanage clear mypool/mydataset

# Skip confirmation prompt with --force flag
tnmanage clear mypool/mydataset --force

# Using flags
tnmanage clear mypool/mydataset --server https://10.0.5.30 --token your-api-token
```

**Warning:** This operation deletes all contents of the dataset and recreates it empty. This cannot be undone!

## TrueNAS API Key

To create an API key in TrueNAS:
1. Go to Settings → API Keys
2. Click "Add"
3. Give it a name and save
4. Copy the generated API key

